package ringio

import (
	"github.com/ciiim/cloudborad/ringio/fspb"
	"github.com/ciiim/cloudborad/storage/hashchunk"
	"github.com/ciiim/cloudborad/storage/tree"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func PBFileInfoToHashFileInfo(pb *fspb.HashChunkInfo) *hashchunk.HashChunkInfo {
	if pb == nil {
		return nil
	}
	return &hashchunk.HashChunkInfo{
		// 计数为-1代表为远程chunk文件
		ChunkCount:   -1,
		ChunkName:    pb.ChunkName,
		ChunkHash:    pb.ChunkHash,
		ChunkPath:    pb.BasePath,
		ChunkSize:    pb.Size,
		ChunkModTime: pb.ModTime.AsTime(),
	}
}

func PbSubsToSubs(pb []*fspb.SubInfo) []*tree.SubInfo {
	subs := make([]*tree.SubInfo, len(pb))
	for i, v := range pb {
		subs[i] = &tree.SubInfo{
			Name:    v.Name,
			IsDir:   v.IsDir,
			ModTime: v.ModTime.AsTime(),
		}
	}
	return subs
}

func SubsToPbSubs(subs []*tree.SubInfo) []*fspb.SubInfo {
	pbSubs := make([]*fspb.SubInfo, len(subs))
	for i, v := range subs {
		pbSubs[i] = &fspb.SubInfo{
			Name:    v.Name,
			IsDir:   v.IsDir,
			ModTime: timestamppb.New(v.ModTime),
		}
	}
	return pbSubs
}

func WithPort(addr string, port string) string {
	return addr + ":" + port
}
