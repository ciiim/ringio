package server

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ciiim/cloudborad/internal/fs"
	"github.com/ciiim/cloudborad/internal/fs/peers"
)

//TODO: 事务(transaction)系统，提供事务接口，支持事务回滚，实现文件下载、上传、删除的事务
//TODO: 秒传模块，相同Hash的文件秒传 DONE
//TODO: 节点强一致性，保证节点信息一致性

type storeBlocks struct {
	storeID    string
	space      string
	base       string
	fileName   string
	hash       string
	blocks     [][]byte
	blockDatas []fs.Fileblock
	nums       int
	now        int
}

type Server struct {
	serverName string
	_IP        string
	Group      *fs.Group

	mu       sync.RWMutex
	storeMap map[string]*storeBlocks
}

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

分片上传文件步骤 （文件不允许大于?GB）
1.BeginStoreFile 此时检查文件是否存在，如果存在则返回已存在的文件信息，
否则创建一个全局唯一的标识用于后续接受文件分片，还有一个切片用于存储文件分片。

2.StoreBlock 加入文件分片至切片。

3.EndStoreFile 创建文件元数据，将文件元数据存储至前端文件系统，将文件切片存储至文件系统。

*/

func (s *Server) BeginStoreFile(space, base, name, hash string, blocksNum int) (storeID string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	storeID = genStoreID()
	s.storeMap[storeID] = &storeBlocks{
		storeID:    storeID,
		space:      space,
		base:       base,
		fileName:   name,
		hash:       hash,
		blocks:     make([][]byte, blocksNum),
		blockDatas: make([]fs.Fileblock, blocksNum),
		nums:       blocksNum,
		now:        0,
	}
	return
}

func (s *Server) StoreBlock(storeID, hash string, data []byte) error {
	s.mu.Lock()
	info, ok := s.storeMap[storeID]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("storeID not exist")
	}
	if info.now > info.nums {
		s.mu.Lock()
		delete(s.storeMap, storeID)
		s.mu.Unlock()
		return fmt.Errorf("block nums is enough,something wrong")
	}
	info.blocks[info.now] = data
	info.blockDatas[info.now] = fs.NewFileBlock(s.Group.StoreSystem.Peer().Pick(hash).PAddr(), int64(len(data)), hash)
	info.now++
	return nil
}

func (s *Server) EndStoreFile(storeID string) error {
	s.mu.RLock()
	info, ok := s.storeMap[storeID]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("storeID not exist")
	}
	for i, v := range info.blocks {
		bi := info.blockDatas[i]
		s.Group.StoreSystem.Store(bi.Hash, fmt.Sprintf("%s_%d", info.fileName, bi.BlockID), v)
	}
	metadata := fs.NewMetaData(info.fileName, info.hash, time.Now(), info.blockDatas)
	data, _ := fs.MarshalMetaData(&metadata)
	err := s.Group.FrontSystem.PutMetadata(info.space, info.base, info.fileName, info.hash, data)
	s.mu.Lock()
	delete(s.storeMap, storeID)
	s.mu.Unlock()
	return err
}

func (s *Server) GetFile(space, base, name string) {

}

func (s *Server) DeleteFile(space, base, name string) {

}

func (s *Server) MakeDir(space, base, name string) error {
	return s.Group.FrontSystem.MakeDir(space, base, name)
}

func (s *Server) RenameDir(space, base, name, newName string) error {
	return s.Group.FrontSystem.RenameDir(space, base, name, newName)
}

func (s *Server) DeleteDir(space, base, name string) error {
	return s.Group.FrontSystem.DeleteDir(space, base, name)
}

func (s *Server) GetDirSub(space, base, name string) ([]fs.SubInfo, error) {
	return s.Group.FrontSystem.GetDirSub(space, base, name)
}

func (s *Server) NewBoard(space string) error {
	return s.Group.FrontSystem.NewSpace(space, fs.GB)
}

func (s *Server) DeleteBoard(space string) error {
	return s.Group.FrontSystem.DeleteSpace(space)
}

func (s *Server) JoinCluster(clusterMemberPeer peers.PeerInfo) error {
	//transaction start
	err := s.Group.FrontSystem.Peer().PActionTo(peers.P_ACTION_JOIN, clusterMemberPeer)
	if err != nil {
		//rollback
		return err
	}
	err = s.Group.StoreSystem.Peer().PActionTo(peers.P_ACTION_JOIN, clusterMemberPeer)
	if err != nil {
		//rollback
		return err
	}
	list, err := s.Group.FrontSystem.Peer().GetPeerListFromPeer(clusterMemberPeer)
	if err != nil {
		//rollback
		return err
	}
	//transaction end
	s.Group.FrontSystem.Peer().PAdd(list...)
	s.Group.StoreSystem.Peer().PAdd(list...)
	return nil

}

func (s *Server) QuitCluster() error {
	list := s.Group.FrontSystem.Peer().PList()

	err := s.Group.FrontSystem.Peer().PActionTo(peers.P_ACTION_QUIT, list...)
	if err != nil {
		return err
	}
	return s.Group.StoreSystem.Peer().PActionTo(peers.P_ACTION_QUIT, list...)

}
