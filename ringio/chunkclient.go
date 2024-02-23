package ringio

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ciiim/cloudborad/chunkpool"
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
	RPCBufferSize int64
	pool          *chunkpool.ChunkPool
}

func newRPCHashClient(pool *chunkpool.ChunkPool) *rpcHashClient {
	return &rpcHashClient{
		RPCBufferSize: DefaultRPCBufferSize,
		pool:          pool,
	}
}

func ctxWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), _RPC_TIMEOUT)
}

func (c *rpcHashClient) dialClient(ctx context.Context, ni *node.Node) (fspb.HashChunkSystemServiceClient, func(), error) {
	conn, err := grpc.DialContext(ctx, ni.Addr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return fspb.NewHashChunkSystemServiceClient(conn), func() {
		conn.Close()
	}, nil
}

type tempFileReadSeekCloser struct {
	*os.File
}

func (t *tempFileReadSeekCloser) Close() error {
	err := t.File.Close()
	os.Remove(t.File.Name())
	return err
}

func warpTempFileReadSeekCloser(file *os.File) io.ReadSeekCloser {
	return &tempFileReadSeekCloser{file}
}

func (c *rpcHashClient) get(ctx context.Context, ni *node.Node, key []byte) (chunk *DHashChunk, err error) {
	defer func(err *error) {
		if *err != nil {
			*err = errors.New("remote get chunk: " + (*err).Error())
		}
	}(&err)

	client, close, err := c.dialClient(ctx, ni)
	if err != nil {
		return nil, err
	}
	defer close()
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

	chunk.HashChunk.SetInfo(hashchunk.NewInfo(PBChunkInfoToHashChunkInfo(resp.ChunkInfo), nil))
	// 如果chunk大小超过默认buffer大小，写入临时文件中
	if resp.ChunkInfo.Size > c.RPCBufferSize {
		chunkTempFile, err := os.CreateTemp(os.TempDir(), "remote-chunk-")
		if err != nil {
			return nil, err
		}
		//不要defer关闭，接受完数据就seek到文件头，然后返回
		defer func(err *error) {
			// 如果err不为nil，说明在接受chunk数据时出现了错误，需要删除临时文件
			if *err != nil {
				if cerr := chunkTempFile.Close(); cerr != nil {
					*err = cerr
				}
				os.Remove(chunkTempFile.Name())
			}
		}(&err)

		receivedBuffer := 0

		// 接受chunk数据
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}

			_, err = chunkTempFile.Write(resp.GetData())
			if err != nil {
				return nil, err
			}
			receivedBuffer += len(resp.Data)
		}

		if err = stream.CloseSend(); err != nil {
			return nil, err
		}

		if _, err = chunkTempFile.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}

		if receivedBuffer != int(resp.ChunkInfo.Size) {
			return nil, fmt.Errorf("received size not match: %d != expected %d", receivedBuffer, resp.ChunkInfo.Size)
		}

		chunk.ReadSeekCloser = warpTempFileReadSeekCloser(chunkTempFile)

		return chunk, nil
	} else {
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

		if err = stream.CloseSend(); err != nil {
			return nil, err
		}

		chunk.ReadSeekCloser = chunkBuffer.ReadCloser(c.pool)
		return chunk, nil
	}
}

func (c *rpcHashClient) put(ctx context.Context, ni *node.Node, key []byte, size int64, chunkName string, reader io.Reader) (err error) {
	defer func(err *error) {
		if *err != nil {
			*err = errors.New("remote put chunk: " + (*err).Error())
		}
	}(&err)

	client, close, err := c.dialClient(ctx, ni)
	if err != nil {
		return err
	}
	defer close()

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
	content.ChunkSize = size

	if err = stream.Send(content); err != nil {
		return err
	}

	content = new(fspb.PutRequest)

	transfered := 0

	buffer := make([]byte, c.RPCBufferSize)
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
		if buffered < c.RPCBufferSize {
			continue
		}
		content.Data = buffer[:buffered]
		if err = stream.Send(content); err != nil {
			return err
		}
		transfered += int(buffered)
		buffered = 0
	}

	if buffered != 0 {
		content.Data = buffer[:buffered]
		if err = stream.Send(content); err != nil {
			return err
		}
		transfered += int(buffered)
	}

	if transfered != int(size) {
		return errors.New("transfered size not match")
	}

	println("transfered size: ", transfered)

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}

	if resp.GetErr() != "" {
		return errors.New(resp.GetErr())
	}

	return nil
}

func (c *rpcHashClient) delete(ctx context.Context, ni *node.Node, key []byte) (err error) {
	defer func(err *error) {
		if *err != nil {
			*err = errors.New("remote delete chunk: " + (*err).Error())
		}
	}(&err)
	client, close, err := c.dialClient(ctx, ni)
	if err != nil {
		return err
	}
	defer close()

	resp, err := client.Delete(ctx, &fspb.Key{Key: key})
	if err != nil {
		return err
	}

	if resp.GetErr() != "" {
		return err
	}

	return nil
}
