// implement basic file system
package fs

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"time"

	"github.com/ciiim/cloudborad/internal/database"
	"github.com/ciiim/cloudborad/internal/fs/peers"

	"github.com/syndtr/goleveldb/leveldb"
)

type basicFileSystem struct {
	rootPath string //相对路径 relative path
	capacity Byte
	occupy   Byte

	fileInfoDBName string

	levelDB *leveldb.DB //concurrent safe

	calcStoreFilePathFn CalcStoreFilePathFnType

	HashFn Hash
}

type CalcStoreFilePathFnType = func(fileinfo BasicFileInfo) string

type Hash func([]byte) string

type BasicFile struct {
	data []byte
	info BasicFileInfo
}

type BasicFileInfo struct {
	FileName string    `json:"fileName"`
	Hash_    string    `json:"hash"`
	Path_    string    `json:"path"`
	Size_    int64     `json:"size"`
	Dir_     bool      `json:"dir"`
	ModTime_ time.Time `json:"modTime"`
}

// default calculate store path function
// format: year/month/day/filehash[0:3]/filehash[3:6]
var DefaultCalcStorePathFn = func(bfi BasicFileInfo) string {
	path := ""
	timePath := time.Time.Format(time.Now(), "2006/01/02")
	path = fmt.Sprintf("%s/%s/%s", timePath, bfi.Hash_[0:3], bfi.Hash_[3:6])
	return path
}

var DefaultHashFn Hash = func(b []byte) string {
	return fmt.Sprintf("%x", b)
}

var _ FileSystem = (*basicFileSystem)(nil)
var _ FileInfo = (*BasicFileInfo)(nil)

func newBasicFileSystem(rootPath string, capacity int64, calcStorePathFn CalcStoreFilePathFnType) *basicFileSystem {
	if err := os.MkdirAll(rootPath, os.ModePerm); err != nil {
		panic("mkdir error:" + err.Error())
	}
	hashDBName := "fileinfo_hash"
	db, err := database.NewLevelDB(rootPath + "/" + hashDBName)
	if err != nil {
		panic("leveldb init error:" + err.Error())
	}

	bfs := &basicFileSystem{
		rootPath:            rootPath,
		capacity:            capacity,
		fileInfoDBName:      hashDBName,
		levelDB:             db,
		calcStoreFilePathFn: calcStorePathFn,
	}
	if calcStorePathFn == nil {
		log.Println("[BFS] Use Default Calculate Function.")
		bfs.calcStoreFilePathFn = DefaultCalcStorePathFn
	}

	cap, ouppy, err := getCapAndOccupy(bfs.levelDB)

	if err != nil {
		return bfs
	}
	log.Printf("Detect exist filesystem at %s\n", rootPath)

	bfs.capacity = cap
	bfs.occupy = ouppy

	if capacity < cap {
		log.Println("[BFS] capacity is less than exist filesystem, use exist filesystem's capacity.")
	}
	if capacity > cap {
		log.Println("[BFS] capacity is more than exist filesystem, use new capacity.")
		bfs.capacity = capacity
	}
	return bfs
}

func (bfs *basicFileSystem) Store(key, fileName string, value []byte) error {
	if key == "" {
		return fmt.Errorf("key is empty")
	}
	if value == nil {
		return fmt.Errorf("value is nil")
	}

	//check exist
	if bfs.isExist(key) {
		return nil //ErrExist //XXX: 需要一个更好的处理方案
	}
	//check capacity
	if bfs.occupy+int64(len(value)) > bfs.capacity {
		return ErrFull
	}

	bfi := NewFileInfo(fileName, key, "", int64(len(value)), false)

	// bfi.Path = rootPath/<path>
	bfi.Path_ = bfs.rootPath + "/" + bfs.calcStoreFilePathFn(bfi)
	if bfi.Path_ == "" {
		return fmt.Errorf("CalcStoreFilePathFn error")
	}

	//make dir
	if err := os.MkdirAll(bfi.Path_, os.ModePerm); err != nil {
		return err
	}
	if err := bfs.storeFileInfo(key, bfi); err != nil {
		return err
	}
	if err := bfs.storeFile(bfi, value); err != nil {
		return err
	}

	//update occupy
	bfs.occupy += bfi.Size_

	return nil
}

func (bfs *basicFileSystem) Get(key string) (File, error) {
	if key == "" {
		return nil, fmt.Errorf("key is empty")
	}
	bfi, err := bfs.getFileInfo(key)
	if err != nil {
		return nil, err
	}
	data, err := bfs.getFile(bfi)
	return BasicFile{
		data: data,
		info: bfi,
	}, err
}

func (bfs *basicFileSystem) Delete(key string) error {
	if key == "" {
		return fmt.Errorf("key is empty")
	}
	bfi, err := bfs.getFileInfo(key)
	if err != nil {
		return err
	}
	if err := bfs.deleteFileInfo(key); err != nil {
		return err
	}
	if err := bfs.deleteFile(bfi); err != nil {
		return err
	}
	if bfs.occupy == 0 {
		panic("[Delete Panic] occupy is 0")
	}
	//update occupy
	bfs.occupy -= bfi.Size_
	return nil
}

func (bfs *basicFileSystem) Set(opt any) error {
	return nil
}

func (bfs *basicFileSystem) isExist(key string) bool {
	if key == "" {
		return false
	}
	_, err := bfs.getFileInfo(key)
	return err == nil
}

// unit can be "B", "KB", "MB", "GB" or just leave it blank
func (bfs *basicFileSystem) Occupy(unit ...string) float64 {
	if len(unit) == 0 {
		return float64(bfs.occupy)
	}
	switch unit[0] {
	case "B":
		return float64(bfs.occupy)
	case "KB":
		return float64(bfs.occupy) / 1024
	case "MB":
		return float64(bfs.occupy) / 1024 / 1024
	case "GB":
		return float64(bfs.occupy) / 1024 / 1024 / 1024
	default:
		return float64(bfs.occupy)
	}
}

func (bfs *basicFileSystem) storeFile(key BasicFileInfo, value []byte) error {
	if bfs.calcStoreFilePathFn == nil {
		panic("calcStoreFilePathFn is nil")
	}
	file, err := os.Create(key.Path_ + "/" + key.FileName)
	if err != nil {
		return fmt.Errorf("open file %s error: %s", key.Path_+"/"+key.FileName, err)
	}
	defer file.Close()
	_, err = file.Write(value)
	return err
}

func (bfs *basicFileSystem) getFile(key BasicFileInfo) ([]byte, error) {
	path := key.Path_
	defer func() {
		if err := recover(); err != nil {
			log.Println("Get panic:", err)
			path = bfs.rootPath + "/" + "default"
		}
	}()
	if path == "" {
		panic("path is empty")
	}
	file, err := os.Open(path + "/" + key.FileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fi, _ := file.Stat()
	data := make([]byte, fi.Size())
	_, err = file.Read(data)
	return data, err
}

func (bfs *basicFileSystem) deleteFile(bfi BasicFileInfo) error {
	fullPath := bfi.Path_ + "/" + bfi.FileName
	if fullPath == "" {
		return fmt.Errorf("path is empty")
	}
	err := os.Remove(fullPath)
	return err
}

func (bfs *basicFileSystem) getFileInfo(hashSum string) (BasicFileInfo, error) {
	infoBytes, err := bfs.levelDB.Get([]byte(hashSum), nil)
	if err != nil {
		return BasicFileInfo{}, err
	}
	var info BasicFileInfo
	// pbfi := &pb.FileInfo{}
	// err = proto.Unmarshal(infoBytes, pbfi)
	err = json.Unmarshal(infoBytes, &info)
	return info, err
}

func (bfs *basicFileSystem) storeFileInfo(hashSum string, file BasicFileInfo) error {
	if file.FileName == "" {
		return ErrFileInvalidName
	}
	// pbfi := &pb.FileInfo{}
	// basicFileInfoToPBFileInfo(file, pbfi)
	// res, err := proto.Marshal(pbfi)
	res, err := json.Marshal(file)
	if err != nil {
		return err
	}
	err = bfs.levelDB.Put([]byte(hashSum), res, nil)
	return err
}

func (bfs *basicFileSystem) deleteFileInfo(hashSum string) error {
	return bfs.levelDB.Delete([]byte(hashSum), nil)
}

func storeCapAndOccupy(levelDB *leveldb.DB, capacity, occupy int64) error {
	if levelDB == nil {
		panic("levelDB is nil")
	}
	var capAndOccupy struct {
		Capacity int64 `json:"capacity"`
		Occupy   int64 `json:"occupy"`
	}
	capAndOccupy.Capacity = capacity
	capAndOccupy.Occupy = occupy
	res, err := json.Marshal(capAndOccupy)
	if err != nil {
		return err
	}
	err = levelDB.Put([]byte("cap_and_occupy"), res, nil)
	return err
}

func getCapAndOccupy(levelDB *leveldb.DB) (int64, int64, error) {
	if levelDB == nil {
		panic("levelDB is nil")
	}
	res, err := levelDB.Get([]byte("cap_and_occupy"), nil)
	if err != nil {
		return 0, 0, err
	}
	var capAndOccupy struct {
		Capacity int64 `json:"capacity"`
		Occupy   int64 `json:"occupy"`
	}
	err = json.Unmarshal(res, &capAndOccupy)
	if err != nil {
		return 0, 0, err
	}
	return capAndOccupy.Capacity, capAndOccupy.Occupy, nil
}

func (bfs *basicFileSystem) Close() error {
	if bfs == nil {
		panic("basicFileSystem is nil")
	}
	log.Println("basicFileSystem Closing.")

	//save cap and ouppy
	if err := storeCapAndOccupy(bfs.levelDB, bfs.capacity, bfs.occupy); err != nil {
		log.Println("Save filesystem error:", err)
	}
	return bfs.levelDB.Close()
}

func (b BasicFile) Data() []byte {
	return b.data
}

func (b BasicFile) Stat() FileInfo {
	return b.info
}

func NewFileInfo(fileName string, hashSum string, path string, size int64, isDir bool, modTime ...time.Time) BasicFileInfo {

	bfi := BasicFileInfo{
		FileName: fileName,
		Hash_:    hashSum,
		Path_:    path,
		Size_:    size,
		Dir_:     isDir,
	}
	if len(modTime) == 1 {
		bfi.ModTime_ = modTime[0]
	}
	return bfi
}

func (bfi BasicFileInfo) Name() string {
	return bfi.FileName
}

func (bfi BasicFileInfo) Path() string {
	return bfi.Path_
}

func (bfi BasicFileInfo) Hash() string {
	return bfi.Hash_
}

func (bfi BasicFileInfo) Size() int64 {
	return int64(bfi.Size_)
}

func (bfi BasicFileInfo) IsDir() bool {
	return bfi.Dir_
}

func (bfi BasicFileInfo) ModTime() time.Time {
	return time.Time{}
}

func (bfi BasicFileInfo) Mode() os.FileMode {
	return 0
}

func (BasicFileInfo) SubDir() []fs.DirEntry {
	return nil
}

func (bfi BasicFileInfo) PeerInfo() peers.PeerInfo {
	return peers.LocalPeerInfo{}
}
