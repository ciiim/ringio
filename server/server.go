package server

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/ciiim/cloudborad/internal/fs"
	"github.com/ciiim/cloudborad/internal/fs/peers"
)

type BeginStoreInfo struct {
	hash string
}

type storeBlocks struct {
	storeID    int64
	blocks     [][]byte
	blockDatas []fs.Fileblock
	nums       int
	now        int
}

type Server struct {
	serverName string
	_IP        string
	Group      *fs.Group

	storeMap map[int64]*storeBlocks
}

/*
ffs is the front file system

it must be a tree structure
*/
func NewServer(groupName, serverName, addr string) *Server {
	if addr == "" {
		addr = GetIP()
	}
	log.Printf("[Server] New server <%s>-<%s>", serverName, addr)
	ffs := fs.NewTreeDFileSystem(*fs.NewDPeer("front0_"+serverName+"_"+groupName, fs.WithPort(addr, fs.RPC_TDFS_PORT), 20, nil), "./front0_"+serverName+"_"+groupName)
	sfs := fs.NewDFS(*fs.NewDPeer("store0_"+serverName+"_"+groupName, fs.WithPort(addr, fs.RPC_HDFS_PORT), 20, nil), "./store0_"+serverName+"_"+groupName, 50*fs.GB, nil)
	if ffs == nil || sfs == nil {
		log.Fatal("New server failed")
	}
	server := &Server{
		Group:      fs.NewGroup(groupName, ffs, sfs),
		serverName: serverName,
		_IP:        addr,
	}
	return server
}

func (s *Server) StartServer(addr string, apiServiceEnable bool) {
	s.Group.Serve()
}

/*

分片上传文件步骤 （文件不允许大于1G）
1.BeginStoreFile 此时检查文件是否存在，如果存在则返回已存在的文件信息，
否则创建一个全局唯一的标识用于后续接受文件分片，还有一个切片用于存储文件分片。

2.StoreBlock 加入文件分片至切片。

3.EndStoreFile 创建文件元数据，将文件元数据存储至前端文件系统，将文件切片存储至文件系统。

*/

func (s *Server) BeginStoreFile(space, base, name, hash string, blocksNum int) (storeID int64, err error) {
	timeStr := strconv.Itoa(int(time.Now().UnixMilli()))
	sum := sha1.Sum([]byte(timeStr))
	fmt.Printf("hex.EncodeToString(sum[:]): %v\n", hex.EncodeToString(sum[:]))
	return 0, nil
}

func (s *Server) StoreBlock(space, base, name, hash string, data []byte) {

}

func (s *Server) EndStoreFile() {

}

func (s *Server) GetFile(space, base, name string) {

}

func (s *Server) DeleteFile(space, base, name string) {

}

func (s *Server) MakeDir(space, base, name string) {

}

func (s *Server) RenameDir(space, base, name, newName string) {

}

func (s *Server) DeleteDir(space, base, name string) {

}

func (s *Server) GetDirSub(space, base, name string) {

}

func (s *Server) NewBoard(space string) {

}

func (s *Server) DeleteBoard(space string) {

}

func (s *Server) JoinCluster(clusterMemberPeer peers.PeerInfo) {

}

func (s *Server) QuitCluster() {

}
