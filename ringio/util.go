package ringio

import (
	"github.com/ciiim/cloudborad/replica"
	"github.com/ciiim/cloudborad/ringio/fspb"
	"github.com/ciiim/cloudborad/storage/hashchunk"
	"github.com/ciiim/cloudborad/storage/tree"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func HashChunkInfoToPBChunkInfo(info *hashchunk.HashChunkInfo) *fspb.HashChunkInfo {
	if info == nil {
		return nil
	}
	return &fspb.HashChunkInfo{
		ChunkCount: info.ChunkCount,
		ChunkName:  info.ChunkName,
		ChunkHash:  info.ChunkHash,
		BasePath:   info.ChunkPath,
		Size:       info.ChunkSize,
		ModTime:    timestamppb.New(info.ChunkModTime),
		CreateTime: timestamppb.New(info.ChunkCreateTime),
	}
}

func PBChunkInfoToHashChunkInfo(pb *fspb.HashChunkInfo) *hashchunk.HashChunkInfo {
	if pb == nil {
		return nil
	}
	return &hashchunk.HashChunkInfo{
		ChunkCount:   pb.ChunkCount,
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

func ReplicaInfoToPBReplicaInfo(chunkInfo *hashchunk.HashChunkInfo, info *replica.ReplicaObjectInfo) *fspb.ReplicaChunkInfo {
	if info == nil {
		return nil
	}
	return &fspb.ReplicaChunkInfo{
		ChunkInfo:    HashChunkInfoToPBChunkInfo(chunkInfo),
		Master:       info.Master,
		ReplicaCount: int64(info.ReplicaCount),
		Checksum:     info.Checksum,
		NodeIds:      info.All,
		Custom:       info.Custom,
	}
}

func PBReplicaInfoToReplicaInfo(pb *fspb.ReplicaChunkInfo) (*hashchunk.HashChunkInfo, *replica.ReplicaObjectInfo) {
	if pb == nil {
		return nil, nil
	}
	return PBChunkInfoToHashChunkInfo(pb.ChunkInfo), &replica.ReplicaObjectInfo{
		Master:       pb.Master,
		ReplicaCount: int(pb.ReplicaCount),
		Checksum:     pb.Checksum,
		All:          pb.NodeIds,
		Custom:       pb.Custom,
	}
}
