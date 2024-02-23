package ringio

import (
	"context"
	"errors"
	"io"
	"slices"
	"time"

	"github.com/ciiim/cloudborad/replica"
	"github.com/ciiim/cloudborad/ringio/fspb"
	"github.com/ciiim/cloudborad/storage/hashchunk"
)

// 直接存在本地
func (r *rpcServer) PutReplica(stream fspb.HashChunkSystemService_PutReplicaServer) error {
	req, err := stream.Recv()
	if err != nil {
		return err
	}

	replicaInfo := PBReplicaInfoToReplicaInfo(req.Info)
	chunkInfo := replicaInfo.Custom
	if chunkInfo == nil {
		return errors.New("no chunk info")
	}
	replicaInfo.ClearCustom()
	//新建一个chunk
	w, err := r.hcs.local().CreateChunk(chunkInfo.ChunkHash, chunkInfo.ChunkName, chunkInfo.ChunkSize, hashchunk.NewExtraInfo("replica", replicaInfo))
	if err != nil {
		stream.SendAndClose(&fspb.Error{Operation: "New Chunk", Err: err.Error()})
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
			stream.SendAndClose(&fspb.Error{Operation: "Receiving Chunk", Err: err.Error()})
			return nil
		}
		if _, err = w.Write(req.GetData()); err != nil {
			stream.SendAndClose(&fspb.Error{Operation: "Writing Chunk", Err: err.Error()})
			return nil
		}
	}
	stream.SendAndClose(&fspb.Error{})
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
	replicaInfoG, ok := replicaInfo.(*replica.ReplicaObjectInfoG[*hashchunk.HashChunkInfo])
	if !ok {
		return replica.ErrReplicaInfoNotFound
	}

	replicaInfoG.Set(chunk.Info().ChunkInfo)
	pbInfo := ReplicaInfoToPBReplicaInfo(replicaInfoG)

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
		err := s.hcs.local().Delete(key.GetKey())
		if err != nil {
			errCh <- err
			return
		}
		deleted <- struct{}{}
	}()

	select {
	case <-deleted:
		return nil, nil
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
		localReplicaInfoG, ok := localReplicaInfo.(*replica.ReplicaObjectInfoG[*hashchunk.HashChunkInfo])
		if !ok {
			return &fspb.Error{Operation: "Check Replica", Err: replica.ErrReplicaInfoNotFound.Error()}, nil
		}

		if localReplicaInfoG.Count() != remoteReplicaInfo.Count() {
			return &fspb.Error{Operation: "Check Replica", Err: replica.ErrReplicaInfoCountMismatch.Error()}, nil
		}

		if !slices.Equal(localReplicaInfoG.All, remoteReplicaInfo.All) {
			return &fspb.Error{Operation: "Check Replica", Err: replica.ErrReplicaInfoAllNodesMismatch.Error()}, nil
		}

		if !slices.Equal(localReplicaInfoG.Checksum, remoteReplicaInfo.Checksum) {
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
		LocalreplicaInfoG, ok := replicaInfo.(*replica.ReplicaObjectInfoG[*hashchunk.HashChunkInfo])
		if !ok {
			errCh <- &fspb.Error{Operation: "Update Replica Info", Err: replica.ErrReplicaInfoNotFound.Error()}
			return
		}

		remoteReplicaInfo := PBReplicaInfoToReplicaInfo(remoteInfo)

		//校验key是否相同
		if !slices.Equal(LocalreplicaInfoG.Key, remoteReplicaInfo.Key) {
			errCh <- &fspb.Error{Operation: "Update Replica Info", Err: replica.ErrReplicaInfoKeyMismatch.Error()}
			return
		}

		//更新本地副本信息
		info.ChunkInfo = remoteReplicaInfo.Custom
		remoteReplicaInfo.ClearCustom()
		info.ExtraInfo.Extra = remoteReplicaInfo

		if err := s.hcs.local().UpdateInfo(LocalreplicaInfoG.Key, func(oldInfo *hashchunk.Info) {
			oldInfo = info
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
