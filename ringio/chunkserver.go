package ringio

import (
	"context"
	"io"

	"github.com/ciiim/cloudborad/ringio/fspb"
	"github.com/ciiim/cloudborad/storage/hashchunk"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (r *rpcServer) Get(key *fspb.Key, stream fspb.HashChunkSystemService_GetServer) error {
	if err := stream.RecvMsg(key); err != nil {
		_ = stream.Send(&fspb.GetResponse{
			Error: &fspb.Error{Operation: "Get Chunk", Err: err.Error()},
		})
		return nil
	}

	chunk, err := r.hcs.local().Get(key.Key)
	if err != nil {
		_ = stream.Send(&fspb.GetResponse{
			Error: &fspb.Error{Operation: "Get Chunk", Err: err.Error()},
		})
		return nil
	}
	fi := chunk.Info().ChunkInfo
	if err = stream.Send(&fspb.GetResponse{
		ChunkInfo: &fspb.HashChunkInfo{
			ChunkCount: fi.Count(),
			ChunkName:  fi.Name(),
			ChunkHash:  fi.Hash(),
			BasePath:   fi.Path(),
			Size:       fi.Size(),
			ModTime:    timestamppb.New(fi.ModTime()),
			CreateTime: timestamppb.New(fi.CreateTime()),
		},
	}); err != nil {
		return err
	}

	var bufSize int64
	// 1/16 of the chunk size
	if bufSize = fi.Size() >> 4; r.defaultBufferSize > bufSize {
		// 小于默认缓冲区大小，使用默认缓冲区大小
		bufSize = r.defaultBufferSize
	}

	buffer := make([]byte, bufSize)
	buffered := 0
	for {
		n, err := chunk.Read(buffer[buffered:])
		if err == io.EOF {
			break
		}
		if err != nil {
			_ = stream.Send(&fspb.GetResponse{
				Error: &fspb.Error{Operation: "Sending Chunk", Err: err.Error()},
			})
			break
		}
		if n == 0 {
			continue
		}
		buffered += n
		if buffered >= int(bufSize) {
			if err = stream.Send(&fspb.GetResponse{
				Data: buffer[:buffered],
			}); err != nil {
				return err
			}
			buffered = 0
		}

	}

	// Send remaining data in buffer
	if buffered > 0 {
		if err = stream.Send(&fspb.GetResponse{
			Data: buffer[:buffered],
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *rpcServer) Put(stream fspb.HashChunkSystemService_PutServer) error {
	request := &fspb.PutRequest{}
	if err := stream.RecvMsg(request); err != nil {
		return nil
	}
	//新建一个chunk，还没生成副本，所以不需要replicaInfo
	w, err := r.hcs.local().CreateChunk(request.Key.GetKey(), request.GetChunkName(), nil)
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

func (r *rpcServer) Delete(ctx context.Context, key *fspb.Key) (*fspb.Error, error) {
	if err := r.hcs.local().Delete(key.Key); err != nil {
		return &fspb.Error{Err: err.Error()}, nil
	}
	return &fspb.Error{}, nil
}
