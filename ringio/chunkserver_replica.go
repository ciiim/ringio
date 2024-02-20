package ringio

import (
	"errors"
	"io"

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
	w, err := r.hcs.local().CreateChunk(chunkInfo.ChunkHash, chunkInfo.ChunkName, hashchunk.NewExtraInfo("replica", replicaInfo))
	if err != nil {
		stream.SendAndClose(&fspb.Error{Operation: "New Chunk", Err: err.Error()})
		return nil
	}
	defer func() {
		hashwc, ok := w.(*hashchunk.HashChunkWriteCloser)
		if !ok {
			w.Close()
			return
		}
		// Flush and close the chunk
		_ = hashwc.Flush()
		hashwc.Close()
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
}
