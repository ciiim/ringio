package ringio

import (
	"context"
	"errors"
	"io"
	"os"
	"time"

	"github.com/ciiim/cloudborad/chunkpool"
	dlogger "github.com/ciiim/cloudborad/debug"
	"github.com/ciiim/cloudborad/node"
	"github.com/ciiim/cloudborad/ringio/fspb"
	"github.com/ciiim/cloudborad/storage/hashchunk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	RPC_FS_PORT  = "9631"
	_RPC_TIMEOUT = time.Second * 5
)

type rpcHashClient struct {
	defaultBufferSize int64
	pool              *chunkpool.ChunkPool
}

func ctxWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), _RPC_TIMEOUT)
}

type tempFileReadCloser struct {
	*os.File
}

func (t *tempFileReadCloser) Read(p []byte) (n int, err error) {
	return t.File.Read(p)
}

func (t *tempFileReadCloser) Close() error {
	err := t.File.Close()
	os.Remove(t.File.Name())
	return err
}

func warpTempFileReadCloser(file *os.File) io.ReadCloser {
	return &tempFileReadCloser{file}
}

func (c *rpcHashClient) get(ctx context.Context, ni *node.Node, key []byte) (chunk *DHashChunk, err error) {
	dlogger.Dlog.LogDebugf("[RPC Client]", "Get from %s", ni.Addr())
	conn, err := grpc.Dial(ni.Addr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := fspb.NewHashChunkSystemServiceClient(conn)
	stream, err := client.Get(ctx, &fspb.Key{Key: key})
	if err != nil {
		return nil, err
	}

	chunk = NewDHashChunk(&hashchunk.HashChunk{})
	// 接受chunk信息
	resp, err := stream.Recv()
	if err != nil {
		return nil, err
	}
	chunk.HashChunk.SetInfo(&hashchunk.HashChunkInfo{
		ChunkCount:   hashchunk.RemoteChunkCount,
		ChunkName:    resp.ChunkInfo.ChunkName,
		ChunkHash:    resp.ChunkInfo.ChunkHash,
		ChunkPath:    resp.ChunkInfo.BasePath,
		ChunkSize:    resp.ChunkInfo.Size,
		ChunkModTime: resp.ChunkInfo.ModTime.AsTime(),
	})

	// 如果chunk大小超过默认buffer大小，写入临时文件中
	if resp.ChunkInfo.Size > c.defaultBufferSize {
		chunkTempFile, err := os.CreateTemp(os.TempDir(), "remote-chunk-")
		if err != nil {
			return nil, err
		}
		//不要defer关闭，接受完数据就seek到文件头，然后返回
		defer func() {
			// 如果err不为nil，说明在接受chunk数据时出现了错误，需要删除临时文件
			if err != nil {
				if cerr := chunkTempFile.Close(); cerr != nil {
					err = cerr
				}
				os.Remove(chunkTempFile.Name())
			}
		}()

		// 接受chunk数据
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}

			n := len(resp.Data)
			if n == 0 {
				continue
			}
			_, err = chunkTempFile.Write(resp.GetData())
			if err != nil {
				return nil, err
			}
		}
		if _, err = chunkTempFile.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}

		chunk.ReadCloser = warpTempFileReadCloser(chunkTempFile)

		return chunk, nil

	} //if end

	chunkBuffer := c.pool.Get()
	// 接受chunk数据
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		n := len(resp.Data)
		if n == 0 {
			continue
		}
		_, err = chunkBuffer.Write(resp.GetData())
		if err == chunkpool.FullBuffer {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	chunk.ReadCloser = chunkBuffer.ReadCloser(c.pool)
	return chunk, nil
}

func (c *rpcHashClient) put(ctx context.Context, ni *node.Node, key []byte, chunkName string, reader io.Reader) error {
	dlogger.Dlog.LogDebugf("[RPC Client]", "Put to %s", ni.Addr())
	conn, err := grpc.Dial(ni.Addr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := fspb.NewHashChunkSystemServiceClient(conn)

	stream, err := client.Put(ctx)
	if err != nil {
		return err
	}
	content := new(fspb.PutRequest)
	// send key and chunk name
	content.Key = &fspb.Key{
		Key: key,
	}
	content.ChunkName = chunkName

	if err = stream.Send(content); err != nil {
		return err
	}

	buffer := make([]byte, 4096)
	var buffered int64 = 0
	for {
		n, err := reader.Read(buffer[buffered:])
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		buffered += int64(n)
		if buffered < c.defaultBufferSize {
			continue
		}
		content.Data = buffer[:buffered]
		if err = stream.Send(content); err != nil {
			return err
		}

		buffered = 0
		clear(buffer)
	}

	if buffered != 0 {
		content.Data = buffer[:buffered]
		if err = stream.Send(content); err != nil {
			return err
		}
	}

	remoteErr, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}
	if remoteErr.GetErr() != "" {
		return errors.New(remoteErr.GetErr())
	}

	return nil
}

func (c *rpcHashClient) delete(ctx context.Context, ni *node.Node, key []byte) error {
	dlogger.Dlog.LogDebugf("[RPC Client]", "Delete file in %s", ni.Addr())
	conn, err := grpc.Dial(ni.Addr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := fspb.NewHashChunkSystemServiceClient(conn)
	_, err = client.Delete(ctx, &fspb.Key{Key: key})
	if err != nil {
		return err
	}
	return nil
}
