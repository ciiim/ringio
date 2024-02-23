package ringio

import (
	"context"
	"fmt"
	"io"

	"github.com/ciiim/cloudborad/ringio/fspb"
)

func (r *rpcServer) Get(key *fspb.Key, stream fspb.HashChunkSystemService_GetServer) error {
	chunk, err := r.hcs.local().Get(key.Key)
	if err != nil {
		_ = stream.Send(&fspb.GetResponse{
			Error: &fspb.Error{Operation: "Get Chunk", Err: err.Error()},
		})
		return err
	}
	fi := chunk.Info().ChunkInfo

	if fi == nil {
		return fmt.Errorf("chunk info is nil")
	}

	if err = stream.Send(&fspb.GetResponse{
		ChunkInfo: HashChunkInfoToPBChunkInfo(fi),
	}); err != nil {
		return err
	}

	sent := 0

	buffer := make([]byte, r.RPCBufferSize)
	var buffered int64 = 0
	for {
		n, err := chunk.Read(buffer[buffered:])
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		buffered += int64(n)
		if buffered < r.RPCBufferSize {
			continue
		}
		if err = stream.Send(&fspb.GetResponse{
			Data: buffer[:buffered],
		}); err != nil {
			return err
		}
		sent += int(buffered)
		buffered = 0
	}

	// Send remaining data in buffer
	if buffered > 0 {
		if err = stream.Send(&fspb.GetResponse{
			Data: buffer[:buffered],
		}); err != nil {
			return err
		}
		sent += int(buffered)
	}

	if sent != int(fi.ChunkSize) {
		return io.ErrUnexpectedEOF
	}

	return nil
}

func (r *rpcServer) Put(stream fspb.HashChunkSystemService_PutServer) (err error) {
	defer func(err *error) {
		if *err != nil {
			fmt.Printf("remote put chunk: %s\n", (*err).Error())
		}
	}(&err)

	request := &fspb.PutRequest{}
	if err := stream.RecvMsg(request); err != nil {
		return err
	}
	//新建一个chunk，还没生成副本，所以不需要replicaInfo
	w, err := r.hcs.local().CreateChunk(request.Key.GetKey(), request.GetChunkName(), request.GetChunkSize(), nil)
	if err != nil {
		return err
	}
	defer func() {
		w.Close()
	}()
	received := 0
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		_, err = w.Write(req.GetData())
		if err != nil {
			return err
		}
		received += len(req.GetData())
	}

	if received != int(request.GetChunkSize()) {
		return io.ErrUnexpectedEOF
	}

	stream.SendAndClose(&fspb.Error{})

	return nil
}

func (r *rpcServer) Delete(ctx context.Context, key *fspb.Key) (resp *fspb.Error, err error) {
	defer func(err *error) {
		if *err != nil {
			fmt.Printf("remote delete chunk: %s\n", (*err).Error())
		}
	}(&err)

	if err := r.hcs.local().Delete(key.Key); err != nil {
		return &fspb.Error{Err: err.Error()}, nil
	}
	return &fspb.Error{}, nil
}
