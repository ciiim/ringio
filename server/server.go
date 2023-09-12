package server

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ciiim/cloudborad/internal/fs"
)

//TODO: 事务(transaction)系统，提供事务接口，支持事务回滚，实现文件下载、上传、删除的事务
//TODO: 秒传模块，相同Hash的文件秒传 DONE
//TODO: 节点强一致性，保证节点信息一致性

type downloadTask struct {
	downloadID string
	fileName   string
	fileSize   int64
	fileHash   string
	metadata   *fs.Metadata
}

type DownloadTaskInfo struct {
	FileName string
	FileSize int64
	FileHash string
}

/*
TODO: 使用本机在群组中的id和文件hash和其他信息进行拼接生成唯一的任务id

	storeTaskID@serverID@filehash@timestamp -> base64 encode -> storeID
*/
type storeBlocks struct {
	storeID        string
	lastUploadTime time.Time
	space          string
	base           string
	fileName       string
	hash           string
	blocks         [][]byte
	blockDatas     []fs.Fileblock
	nums           int
	now            int
}

type Server struct {
	stopChan chan struct{}

	serverName string
	_IP        string
	_Port      string
	Group      *fs.Group

	storeMutex sync.RWMutex
	storeMap   map[string]*storeBlocks

	downloadMutex sync.RWMutex
	downloadMap   map[string]*downloadTask
}

func NewServer(serverName, ip, port string, stopChan chan struct{}, options ...ServerOptions) *Server {
	if ip == "" {
		ip = GetIP()
	}
	if port == "" {
		port = fs.RPC_FS_PORT
	}
	log.Printf("[Server] New server <%s>-<%s>", serverName, ip)
	server := &Server{
		stopChan:    stopChan,
		serverName:  serverName,
		_IP:         ip,
		_Port:       port,
		storeMap:    make(map[string]*storeBlocks),
		downloadMap: make(map[string]*downloadTask),
	}
	server.handleOptions(options...)
	return server
}

func (s *Server) StartServer() {
	go s.CleanUploadTask(10 * time.Minute)
	if s.stopChan == nil {
		s.Group.Serve()
	} else {
		go s.Group.Serve()
		<-s.stopChan
		log.Println("[Server] Closing...")
		s.Group.Close()
	}
}

/*

分片上传文件步骤 （文件不允许大于?GB）
1.BeginStoreFile 此时检查文件是否存在，如果存在则返回已存在的文件信息，
否则创建一个全局唯一的标识用于后续接受文件分片，还有一个切片用于存储文件分片。

2.StoreBlock 加入文件分片至切片。

3.EndStoreFile 创建文件元数据，将文件元数据存储至前端文件系统，将文件切片存储至文件系统。

*/

func (s *Server) BeginStoreFile(space, base, name, hash string, blocksNum int) (storeID string, err error) {
	s.storeMutex.Lock()
	defer s.storeMutex.Unlock()
	storeID = genTaskID(hash, TaskTypeUpload)
	s.storeMap[storeID] = &storeBlocks{
		storeID:        storeID,
		lastUploadTime: time.Now(),
		space:          space,
		base:           base,
		fileName:       name,
		hash:           hash,
		blocks:         make([][]byte, blocksNum),
		blockDatas:     make([]fs.Fileblock, blocksNum),
		nums:           blocksNum,
		now:            0,
	}
	return
}

func (s *Server) StoreBlock(storeID string, index int, hash string, data []byte) error {
	s.storeMutex.RLock()
	info, ok := s.storeMap[storeID]
	s.storeMutex.RUnlock()
	if !ok {
		return fmt.Errorf("storeID not exist")
	}
	if info.now > info.nums {
		s.storeMutex.Lock()
		delete(s.storeMap, storeID)
		s.storeMutex.Unlock()
		return fmt.Errorf("block nums larger than need, upload canceled")
	}
	info.blocks[index] = data
	info.blockDatas[index] = fs.NewFileBlock(s.Group.PeerService.PAddr().String(), int64(len(data)), hash)
	info.lastUploadTime = time.Now()
	info.now++
	if info.now == info.nums {
		log.Println("end store file")
		return s.EndStoreFile(storeID)
	}
	return nil
}

func (s *Server) EndStoreFile(storeID string) error {
	s.storeMutex.RLock()
	info, ok := s.storeMap[storeID]
	s.storeMutex.RUnlock()
	if !ok {
		return fmt.Errorf("storeID not exist")
	}
	if info.now != info.nums {
		return fmt.Errorf("block nums is not enough,something wrong")
	}

	for i, v := range info.blocks {
		bi := info.blockDatas[i]
		s.Group.StoreSystem.Store(bi.Hash, fmt.Sprintf("%s_%s_%d.block", bi.Hash, info.fileName, bi.BlockID), v)
	}
	metadata := fs.NewMetaData(info.fileName, info.hash, time.Now(), info.blockDatas)
	data, _ := fs.MarshalMetaData(&metadata)
	err := s.Group.FrontSystem.PutMetadata(info.space, info.base, info.fileName+".meta", info.hash, data)
	s.storeMutex.Lock()
	delete(s.storeMap, storeID)
	s.storeMutex.Unlock()
	return err
}

func (s *Server) CheckUploadStatus(storeID string) (int, error) {
	s.storeMutex.RLock()
	info, ok := s.storeMap[storeID]
	s.storeMutex.RUnlock()
	if !ok {
		return -1, fmt.Errorf("storeID not exist")
	}
	return info.now, nil
}

func (s *Server) CleanUploadTask(t time.Duration) {
	if s.storeMap == nil {
		return
	}
	log.Println("[Server] clean upload task start.")
	for {
		time.Sleep(t)
		s.storeMutex.Lock()
		for k, v := range s.storeMap {
			if time.Since(v.lastUploadTime) > t {
				delete(s.storeMap, k)
			}
		}
		s.storeMutex.Unlock()
	}

}

func (s *Server) BeginDownloadFile(space, base, name string) (downloadID string, fileSize int64, blockNum int, err error) {
	metadataBytes, err := s.Group.FrontSystem.GetMetadata(space, base, name)
	if err != nil {
		return "", 0, 0, err
	}
	metadata := &fs.Metadata{}
	fs.UnmarshalMetaData(metadataBytes, metadata)
	downloadID = genTaskID(metadata.Hash, TaskTypeDownload)
	blockNum = len(metadata.Blocks)
	fileSize = metadata.Size
	s.downloadMutex.Lock()
	defer s.downloadMutex.Unlock()
	s.downloadMap[downloadID] = &downloadTask{
		downloadID: downloadID,
		fileName:   name,
		fileSize:   metadata.Size,
		fileHash:   metadata.Hash,
		metadata:   metadata,
	}
	return
}

func (s *Server) GetBlock(downloadID string, blockIndex int) ([]byte, error) {
	s.downloadMutex.RLock()
	info, ok := s.downloadMap[downloadID]
	s.downloadMutex.RUnlock()
	if !ok {
		return nil, fmt.Errorf("downloadID not exist")
	}
	if blockIndex >= len(info.metadata.Blocks) {
		return nil, fmt.Errorf("blockIndex out of range")
	}
	block := info.metadata.Blocks[blockIndex]
	file, err := s.Group.StoreSystem.Get(block.Hash)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, fmt.Errorf("block not exist")
	}
	return file.Data(), nil
}

func (s *Server) GetBlockByRange(downloadID string, start, end int64) ([]byte, error) {
	s.downloadMutex.RLock()
	info, ok := s.downloadMap[downloadID]
	s.downloadMutex.RUnlock()
	if !ok {
		return nil, fmt.Errorf("downloadID not exist")
	}
	if start > end || end > info.fileSize {
		return nil, fmt.Errorf("range out of range")
	}
	for i := 0; i < len(info.metadata.Blocks); i++ {
		block := info.metadata.Blocks[i]
		log.Printf("block %d: %d-%d", i, block.Offset, block.Offset+block.Size)
		//Range处于一个Block内
		if block.Offset <= start && block.Offset+end <= block.Size {
			log.Println("range in one block")
			data, err := s.GetBlock(downloadID, i)
			data = data[start-block.Offset : end-block.Offset]
			return data, err
		}
		//Range的start处于当前Block，end处于下一个Block
		if block.Offset <= start && block.Offset+end > block.Size {
			log.Println("range in two block")
			firstData, err := s.GetBlock(downloadID, i)
			if err != nil {
				return nil, err
			}
			secondData, err := s.GetBlock(downloadID, i+1)
			if err != nil {
				return nil, err
			}
			firstData = firstData[start-block.Offset:]
			secondData = secondData[:end-block.Offset]
			return append(firstData, secondData...), nil
		}
		//FIX: Range横跨多个Block的情况
	}
	return nil, fmt.Errorf("range out of range")
}

func (s *Server) DownloadTaskInfo(downloadID string) DownloadTaskInfo {
	s.downloadMutex.RLock()
	info, ok := s.downloadMap[downloadID]
	s.downloadMutex.RUnlock()
	if !ok {
		return DownloadTaskInfo{}
	}
	return DownloadTaskInfo{
		FileName: info.fileName,
		FileSize: info.fileSize,
		FileHash: info.fileHash,
	}
}

func (s *Server) EndDownloadFile(downloadID string) error {
	s.downloadMutex.Lock()
	defer s.downloadMutex.Unlock()
	delete(s.downloadMap, downloadID)
	return nil
}

func (s *Server) DeleteFile(space, base, name string) error {
	//获取文件元数据
	metadataBytes, err := s.Group.FrontSystem.GetMetadata(space, base, name)
	if err != nil {
		log.Printf("[Server] DeleteFile failed: %v", err)
		return err
	}
	metadata := &fs.Metadata{}
	fs.UnmarshalMetaData(metadataBytes, metadata)
	//删除文件分片
	for _, v := range metadata.Blocks {
		if err := s.Group.StoreSystem.Delete(v.Hash); err != nil {
			return err
		}
	}

	//删除文件元数据
	return s.Group.FrontSystem.DeleteMetadata(space, base, name, metadata.Hash)
}

func (s *Server) MakeDir(space, base, name string) error {
	return s.Group.FrontSystem.MakeDir(space, base, name)
}

func (s *Server) RenameDir(space, base, name, newName string) error {
	return s.Group.FrontSystem.RenameDir(space, base, name, newName)
}

func (s *Server) DeleteDir(space, base, name string) error {
	//TODO 删除文件夹时，需要删除文件夹内所有文件
	return s.Group.FrontSystem.DeleteDir(space, base, name)
}

func (s *Server) GetDirSub(space, base, name string) ([]fs.SubInfo, error) {
	return s.Group.FrontSystem.GetDirSub(space, base, name)
}

func (s *Server) NewBoard(space string, cap fs.Byte) error {
	return s.Group.FrontSystem.NewSpace(space, cap)
}

func (s *Server) DeleteBoard(space string) error {
	return s.Group.FrontSystem.DeleteSpace(space)
}
