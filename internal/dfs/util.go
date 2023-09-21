package dfs

import (
	"github.com/ciiim/cloudborad/internal/dfs/fspb"
	"github.com/ciiim/cloudborad/internal/fs"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func PBFileInfoToHashFileInfo(pb *fspb.HashFileInfo) fs.HashFileInfo {
	if pb == nil {
		return fs.HashFileInfo{}
	}
	return fs.HashFileInfo{
		Path_:    pb.BasePath,
		FileName: pb.FileName,
		Hash_:    pb.Hash,
		Size_:    pb.Size,
		ModTime_: pb.ModTime.AsTime(),
	}
}

func PbSubsToSubs(pb []*fspb.SubInfo) []fs.SubInfo {
	var subs []fs.SubInfo
	for _, v := range pb {
		subs = append(subs, fs.SubInfo{
			Name:    v.Name,
			Type:    fs.FILE_TYPE(v.Type),
			ModTime: v.ModTime.AsTime(),
		})
	}
	return subs
}

func SubsToPbSubs(subs []fs.SubInfo) []*fspb.SubInfo {
	var pbSubs []*fspb.SubInfo
	for _, v := range subs {
		pbSubs = append(pbSubs, &fspb.SubInfo{
			Name:    v.Name,
			Type:    string(v.Type),
			ModTime: timestamppb.New(v.ModTime),
		})
	}
	return pbSubs
}

func WithPort(addr string, port string) string {
	return addr + ":" + port
}
