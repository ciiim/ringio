package fs

import (
	"github.com/ciiim/cloudborad/internal/fs/fspb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func pBFileInfoToHashFileInfo(pb *fspb.HashFileInfo) HashFileInfo {
	if pb == nil {
		return HashFileInfo{}
	}
	return HashFileInfo{
		Path_:    pb.BasePath,
		FileName: pb.FileName,
		Hash_:    pb.Hash,
		Size_:    pb.Size,
		ModTime_: pb.ModTime.AsTime(),
	}
}

func pbSubsToSubs(pb []*fspb.SubInfo) []SubInfo {
	var subs []SubInfo
	for _, v := range pb {
		subs = append(subs, SubInfo{
			Name:    v.Name,
			Type:    FILE_TYPE(v.Type),
			ModTime: v.ModTime.AsTime(),
		})
	}
	return subs
}

func subsToPbSubs(subs []SubInfo) []*fspb.SubInfo {
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
