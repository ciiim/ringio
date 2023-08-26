package fs

import (
	"strings"

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

func pbFileInfoToTreeFileInfo(pb *fspb.TreeFileInfo) TreeFileInfo {
	if pb == nil {
		return TreeFileInfo{}
	}
	return TreeFileInfo{
		path:     pb.BasePath,
		fileName: pb.FileName,
		size:     pb.Size,
		modTime:  pb.ModTime.AsTime(),
		isDir:    pb.IsDir,
		subDir:   pbSubsToSubs(pb.SubFiles.SubInfo),
	}
}

func pbSubsToSubs(pb []*fspb.SubInfo) []SubInfo {
	var subs []SubInfo
	for _, v := range pb {
		subs = append(subs, SubInfo{
			Name:    v.Name,
			IsDir:   v.IsDir,
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
			IsDir:   v.IsDir,
			ModTime: timestamppb.New(v.ModTime),
		})
	}
	return pbSubs
}

func WithPort(addr string, port string) string {
	return addr + ":" + port
}

func metadataPathToString3(space, base, name string) string {
	return space + "/\\/" + base + "/\\/" + name
}

func metadataPathToString(m MetadataPath) string {
	return m.Space + "/\\/" + m.Base + "/\\/" + m.Name
}

func stringToMetadataPath3(path string) (string, string, string, bool) {
	s := strings.Split(path, "/\\/")
	if len(s) != 3 {
		return "", "", "", false
	}
	return s[0], s[1], s[2], true
}

func stringToMetadataPath(path string) (MetadataPath, bool) {
	s := strings.Split(path, "/\\/")
	if len(s) != 3 {
		return MetadataPath{}, false
	}
	return MetadataPath{
		Space: s[0],
		Base:  s[1],
		Name:  s[2],
	}, true
}
