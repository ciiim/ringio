package ringio

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/ciiim/cloudborad/replica"
	"github.com/ciiim/cloudborad/ringio/fspb"
	"github.com/ciiim/cloudborad/storage/hashchunk"
	"github.com/ciiim/cloudborad/util"
)

// 直接存在本地
func (r *rpcServer) PutReplica(stream fspb.HashChunkSystemService_PutReplicaServer) (err error) {

	defer func(err *error) {
		if *err != nil {
			fmt.Println(*err)
		}
	}(&err)

	req, err := stream.Recv()
	if err != nil {
		return util.WarpWithDetail(err)
	}

	replicaInfo := PBReplicaInfoToReplicaInfo(req.Info)
	chunkInfo := replicaInfo.Custom
	if chunkInfo == nil {
		return util.WarpWithDetail(errors.New("no chunk info"))
	}
	replicaInfo.Custom = nil
	//新建一个chunk
	w, err := r.hcs.local().CreateChunk(chunkInfo.ChunkHash, chunkInfo.ChunkName, chunkInfo.ChunkSize, hashchunk.NewExtraInfo("replica", replicaInfo))
	if err != nil {
		return util.WarpWithDetail(err)
	}

	// 有相同chunk的情况下，不需要再次写入
	if err == nil && w == nil {
		fmt.Printf("same chunk %x", replicaInfo.Key)
		return nil
	}

	defer func() {
		w.Close()
	}()
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return util.WarpWithDetail(err)
		}
		if _, err = w.Write(req.GetData()); err != nil {
			return util.WarpWithDetail(err)
		}
	}

	if err = stream.SendAndClose(&fspb.Error{}); err != nil {
		return util.WarpWithDetail(err)
	}

	return nil
}

func (s *rpcServer) GetReplica(key *fspb.Key, stream fspb.HashChunkSystemService_GetReplicaServer) error {
	chunk, err := s.hcs.local().Get(key.GetKey())
	if err != nil {
		return err
	}

	defer chunk.Close()

	replicaInfo, ok := chunk.Info().ExtraInfo.TagInfo("replica")
	if !ok {
		return replica.ErrReplicaInfoNotFound
	}

	replicaInfo.Set(chunk.Info().ChunkInfo)
	pbInfo := ReplicaInfoToPBReplicaInfo(replicaInfo)

	content := new(fspb.GetReplicaResponse)

	content.Info = pbInfo

	if err = stream.Send(content); err != nil {
		return err
	}

	content.Info = nil

	var bufSize int64
	// 1/16 of the chunk size
	if bufSize = chunk.Info().ChunkInfo.ChunkSize >> 4; s.RPCBufferSize > bufSize {
		// 小于默认缓冲区大小，使用默认缓冲区大小
		bufSize = s.RPCBufferSize
	}

	buffer := make([]byte, bufSize)
	buffered := 0
	for {
		n, err := chunk.Read(buffer[buffered:])
		if err == io.EOF {
			break
		}
		if err != nil {
			_ = stream.Send(&fspb.GetReplicaResponse{
				Error: &fspb.Error{Operation: "Sending Replica", Err: err.Error()},
			})
			break
		}
		if n == 0 {
			continue
		}
		buffered += n
		if buffered >= int(bufSize) {
			content.Data = buffer[:buffered]
			if err = stream.Send(content); err != nil {
				return err
			}
			buffered = 0
		}
	}

	// Send remaining data in buffer
	if buffered > 0 {
		content.Data = buffer[:buffered]
		if err = stream.Send(content); err != nil {
			return err
		}
	}

	return nil
}

func (s *rpcServer) DeleteReplica(ctx context.Context, key *fspb.Key) (*fspb.Error, error) {
	//XXX: 可能要做一些检查确保不会删错

	deleted := make(chan struct{})
	errCh := make(chan error)
	go func() {
		err := s.hcs.DeleteLocally(key.GetKey())
		if err != nil {
			errCh <- err
			return
		}
		deleted <- struct{}{}
	}()

	select {
	case <-deleted:
		return &fspb.Error{}, nil
	case <-ctx.Done():
		return &fspb.Error{Operation: "Delete Replica", Err: ctx.Err().Error()}, nil
	case <-time.After(_RPC_TIMEOUT):
		return &fspb.Error{Operation: "Delete Replica", Err: "Timeout"}, nil
	case err := <-errCh:
		return &fspb.Error{Operation: "Delete Replica", Err: err.Error()}, nil
	}
}

func (s *rpcServer) CheckReplica(ctx context.Context, req *fspb.CheckReplicaRequest) (*fspb.Error, error) {
	infoCh := make(chan *hashchunk.Info)
	errCh := make(chan error)
	go func() {
		info, err := s.hcs.local().GetInfo(req.Info.Key)
		if err != nil {
			errCh <- err
			return
		}
		infoCh <- info
	}()

	select {
	case <-ctx.Done():
		return &fspb.Error{Operation: "Check Replica", Err: ctx.Err().Error()}, nil
	case <-time.After(_RPC_TIMEOUT):
		return &fspb.Error{Operation: "Check Replica", Err: "Timeout"}, nil
	case err := <-errCh:
		return &fspb.Error{Operation: "Check Replica", Err: err.Error()}, nil
	case info := <-infoCh:
		if info == nil {
			return &fspb.Error{Operation: "Check Replica", Err: replica.ErrReplicaInfoNotFound.Error()}, nil
		}

		remoteReplicaInfo := PBReplicaInfoToReplicaInfo(req.GetInfo())

		localReplicaInfo, ok := info.ExtraInfo.TagInfo("replica")
		if !ok {
			return &fspb.Error{Operation: "Check Replica", Err: replica.ErrReplicaInfoNotFound.Error()}, nil
		}
		if localReplicaInfo.Count() != remoteReplicaInfo.Count() {
			return &fspb.Error{Operation: "Check Replica", Err: replica.ErrReplicaInfoCountMismatch.Error()}, nil
		}

		if !slices.Equal(localReplicaInfo.All, remoteReplicaInfo.All) {
			return &fspb.Error{Operation: "Check Replica", Err: replica.ErrReplicaInfoAllNodesMismatch.Error()}, nil
		}

		if !slices.Equal(localReplicaInfo.Checksum, remoteReplicaInfo.Checksum) {
			return &fspb.Error{Operation: "Check Replica", Err: replica.ErrReplicaInfoChecksumMismatch.Error()}, nil
		}

	}

	return &fspb.Error{}, nil

}

func (s *rpcServer) UpdateReplicaInfo(ctx context.Context, remoteInfo *fspb.ReplicaChunkInfo) (*fspb.Error, error) {
	updated := make(chan struct{})
	errCh := make(chan *fspb.Error)
	go func() {
		info, err := s.hcs.local().GetInfo(remoteInfo.Key)
		if err != nil {
			errCh <- &fspb.Error{Operation: "Update Replica Info", Err: err.Error()}
			return
		}
		if info == nil {
			errCh <- &fspb.Error{Operation: "Update Replica Info", Err: replica.ErrReplicaInfoNotFound.Error()}
			return
		}

		replicaInfo, ok := info.ExtraInfo.TagInfo("replica")
		if !ok {
			errCh <- &fspb.Error{Operation: "Update Replica Info", Err: replica.ErrReplicaInfoNotFound.Error()}
			return
		}
		remoteReplicaInfo := PBReplicaInfoToReplicaInfo(remoteInfo)

		//校验key是否相同
		if !slices.Equal(replicaInfo.Key, remoteReplicaInfo.Key) {
			errCh <- &fspb.Error{Operation: "Update Replica Info", Err: replica.ErrReplicaInfoKeyMismatch.Error()}
			return
		}

		//更新本地副本信息
		info.ChunkInfo = remoteReplicaInfo.Custom
		remoteReplicaInfo.Custom = nil
		info.ExtraInfo.Extra = remoteReplicaInfo

		if err := s.hcs.local().UpdateInfo(remoteReplicaInfo.Key, func(oldInfo *hashchunk.Info) {
			*oldInfo = *info
		}); err != nil {
			errCh <- &fspb.Error{Operation: "Update Replica Info", Err: err.Error()}
			return
		}
		updated <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return &fspb.Error{Operation: "Update Replica Info", Err: ctx.Err().Error()}, nil
	case <-time.After(_RPC_TIMEOUT):
		return &fspb.Error{Operation: "Update Replica Info", Err: "Timeout"}, nil
	case err := <-errCh:
		return err, nil
	case <-updated:
	}

	return &fspb.Error{}, nil
}
