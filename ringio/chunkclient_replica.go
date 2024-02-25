package ringio

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/ciiim/cloudborad/chunkpool"
	"github.com/ciiim/cloudborad/node"
	"github.com/ciiim/cloudborad/replica"
	"github.com/ciiim/cloudborad/ringio/fspb"
	"github.com/ciiim/cloudborad/storage/hashchunk"
	"github.com/ciiim/cloudborad/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (r *rpcHashClient) putReplica(
	ctx context.Context,
	node *node.Node,
	reader io.Reader,
	info *replica.ReplicaObjectInfoG[*hashchunk.HashChunkInfo],
) error {
	conn, err := grpc.DialContext(ctx, node.Addr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return util.WarpWithDetail(err)
	}
	defer conn.Close()

	client := fspb.NewHashChunkSystemServiceClient(conn)
	stream, err := client.PutReplica(ctx)
	if err != nil {
		return util.WarpWithDetail(err)
	}

	content := new(fspb.PutReplicaRequest)

	content.Info = ReplicaInfoToPBReplicaInfo(info)
	if err := stream.Send(content); err != nil {
		return util.WarpWithDetail(err)
	}

	content.Info = nil

	sent := 0

	buffer := make([]byte, r.RPCBufferSize)
	var buffered int64 = 0
	for {
		n, err := reader.Read(buffer[buffered:])
		if err == io.EOF {
			break
		}
		if err != nil {
			return util.WarpWithDetail(err)
		}
		buffered += int64(n)
		if buffered < r.RPCBufferSize {
			continue
		}
		content.Data = buffer[:buffered]
		err = stream.Send(content)
		if err != nil {
			println("sent", sent)
			return util.WarpWithDetail(err)
		}

		sent += int(buffered)

		buffered = 0
	}

	//Flush buffer
	if buffered != 0 {
		content.Data = buffer[:buffered]
		if err = stream.Send(content); err != nil {
			return util.WarpWithDetail(err)
		}
		sent += int(buffered)
	}

	if sent != int(info.Custom.ChunkSize) {
		return util.WarpWithDetail(errors.New("sent size not equal to info size"))
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return util.WarpWithDetail(err)
	}

	if resp.GetErr() != "" {
		return util.WarpWithDetail(errors.New(resp.GetErr()))
	}

	return nil
}

func (r *rpcHashClient) getReplica(ctx context.Context, node *node.Node, key *fspb.Key) (io.ReadSeekCloser, *replica.ReplicaObjectInfoG[*hashchunk.HashChunkInfo], error) {
	client, close, err := r.dialClient(ctx, node)
	if err != nil {
		return nil, nil, util.WarpWithDetail(err)
	}
	defer close()

	stream, err := client.GetReplica(ctx, key)
	if err != nil {
		return nil, nil, util.WarpWithDetail(err)
	}

	resp, err := stream.Recv()
	if err != nil {
		return nil, nil, util.WarpWithDetail(err)
	}

	replicaInfo := PBReplicaInfoToReplicaInfo(resp.Info)
	chunkInfo := replicaInfo.Custom
	// 如果chunk大小超过默认buffer大小，写入临时文件中
	if chunkInfo.Size() > r.RPCBufferSize {
		chunkTempFile, err := os.CreateTemp(os.TempDir(), "remote-chunk-replica-")
		if err != nil {
			return nil, replicaInfo, util.WarpWithDetail(err)
		}
		//不要defer关闭，接受完数据就seek到文件头，然后返回
		defer func(err *error) {
			// 如果err不为nil，说明在接受chunk数据时出现了错误，需要删除临时文件
			if *err != nil && *err != io.EOF {
				if cerr := chunkTempFile.Close(); cerr != nil {
					*err = cerr
				}
				os.Remove(chunkTempFile.Name())
			}
		}(&err)

		// 接受chunk数据
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, replicaInfo, util.WarpWithDetail(err)
			}

			n := len(resp.Data)
			if n == 0 {
				continue
			}
			_, err = chunkTempFile.Write(resp.GetData())
			if err != nil {
				return nil, replicaInfo, util.WarpWithDetail(err)
			}
		}
		if _, err = chunkTempFile.Seek(0, io.SeekStart); err != nil {
			return nil, replicaInfo, util.WarpWithDetail(err)
		}

		rc := warpTempFileReadSeekCloser(chunkTempFile)

		return rc, replicaInfo, nil

	} //if end

	chunkBuffer := r.pool.Get()
	// 接受chunk数据
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, replicaInfo, util.WarpWithDetail(err)
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
			return nil, replicaInfo, util.WarpWithDetail(err)
		}
	}
	rc := chunkBuffer.ReadCloser(r.pool)
	return rc, replicaInfo, nil
}

func (r *rpcHashClient) delReplica(ctx context.Context, node *node.Node, key *fspb.Key) error {
	client, close, err := r.dialClient(ctx, node)
	if err != nil {
		return util.WarpWithDetail(err)
	}
	defer close()

	resp, err := client.DeleteReplica(ctx, key)
	if err != nil {
		return util.WarpWithDetail(err)
	}

	if resp.GetErr() != "" {
		return util.WarpWithDetail(errors.New(resp.GetErr()))
	}

	return nil
}

func (r *rpcHashClient) checkReplica(ctx context.Context, node *node.Node, info *replica.ReplicaObjectInfoG[*hashchunk.HashChunkInfo]) error {
	client, close, err := r.dialClient(ctx, node)
	if err != nil {
		return util.WarpWithDetail(err)
	}
	defer close()

	checkRequest := &fspb.CheckReplicaRequest{
		Info: ReplicaInfoToPBReplicaInfo(info),
	}

	resp, err := client.CheckReplica(ctx, checkRequest)
	if err != nil {
		return util.WarpWithDetail(err)
	}

	if resp.GetErr() != "" {
		return util.WarpWithDetail(errors.New(resp.GetErr()))
	}

	return nil
}

func (r *rpcHashClient) updateReplicaInfo(ctx context.Context, node *node.Node, info *replica.ReplicaObjectInfoG[*hashchunk.HashChunkInfo]) error {
	client, close, err := r.dialClient(ctx, node)
	if err != nil {
		return err
	}
	defer close()

	resp, err := client.UpdateReplicaInfo(ctx, ReplicaInfoToPBReplicaInfo(info))
	if err != nil {
		return err
	}

	if resp.GetErr() != "" {
		return errors.New(resp.GetErr())
	}

	return nil
}
