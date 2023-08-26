// implement hash file system
package fs

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ciiim/cloudborad/internal/database"
	"github.com/ciiim/cloudborad/internal/fs/peers"

	"github.com/syndtr/goleveldb/leveldb"
)

type hashFileSystem struct {
	rootPath string //相对路径 relative path
	capacity Byte
	occupied Byte

	fileInfoDBName string

	levelDB *leveldb.DB

	calcStoreFilePathFn CalcStoreFilePathFnType

	HashFn Hash
}

type CalcStoreFilePathFnType = func(fileinfo HashFileInfo) string

type Hash func([]byte) string

type HashFile struct {
	data []byte
	info HashFileInfo
}

type HashFileInfo struct {
	FileName string    `json:"fileName"`
	Hash_    string    `json:"hash"`
	Path_    string    `json:"path"`
	Size_    int64     `json:"size"`
	ModTime_ time.Time `json:"modTime"`
}

// default calculate store path function
// format: year/month/day/filehash[0:3]/filehash[3:6]
var DefaultCalcStorePathFn = func(bfi HashFileInfo) string {
	path := ""
	timePath := time.Time.Format(time.Now(), "2006/01/02")
	path = fmt.Sprintf("%s/%s/%s", timePath, bfi.Hash_[0:3], bfi.Hash_[3:6])
	return path
}

var DefaultHashFn Hash = func(b []byte) string {
	return fmt.Sprintf("%x", b)
}

var _ HashFileSystemI = (*hashFileSystem)(nil)
var _ HashFileInfoI = (*HashFileInfo)(nil)

func newHashFileSystem(rootPath string, capacity int64, calcStorePathFn CalcStoreFilePathFnType) *hashFileSystem {
	if err := os.MkdirAll(rootPath, os.ModePerm); err != nil {
		panic("mkdir error:" + err.Error())
	}
	hashDBName := "fileinfo_hash"
	db, err := database.NewLevelDB(filepath.Join(rootPath + "/" + hashDBName))
	if err != nil {
		panic("leveldb init error:" + err.Error())
	}

	bfs := &hashFileSystem{
		rootPath:            rootPath,
		capacity:            capacity,
		fileInfoDBName:      hashDBName,
		levelDB:             db,
		calcStoreFilePathFn: calcStorePathFn,
	}
	if calcStorePathFn == nil {
		dlog.debug("[BFS]", "Use Default Calculate Function.")
		bfs.calcStoreFilePathFn = DefaultCalcStorePathFn
	}

	cap, ouppy, err := getCapAndOccupied(bfs.levelDB)

	if err != nil {
		log.Println("New File Storage System at", rootPath)
		storeCapAndOccupied(bfs.levelDB, capacity, 0)
		return bfs
	}
	log.Printf("Detect exist filesystem at %s\n", rootPath)

	bfs.capacity = cap
	bfs.occupied = ouppy

	if capacity < cap {
		log.Println("[BFS] capacity is less than exist filesystem, use exist filesystem's capacity.")
	}
	if capacity > cap {
		log.Println("[BFS] capacity is more than exist filesystem, use new capacity.")
		bfs.capacity = capacity
	}
	return bfs
}

func (bfs *hashFileSystem) Store(key, fileName string, value []byte) error {
	if key == "" {
		return fmt.Errorf("key is empty")
	}
	if value == nil {
		return fmt.Errorf("value is nil")
	}

	//check exist
	if bfs.isExist(key) {
		return ErrFileExist
	}
	//check capacity
	if bfs.occupied+int64(len(value)) > bfs.capacity {
		return ErrFull
	}

	bfi := NewFileInfo(fileName, key, "", int64(len(value)), time.Now())

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

	//update Occupied
	bfs.occupied += bfi.Size_

	return nil
}

func (bfs *hashFileSystem) Get(key string) (HashFileI, error) {
	if key == "" {
		return nil, fmt.Errorf("key is empty")
	}
	bfi, err := bfs.getFileInfo(key)
	if err != nil {
		return nil, err
	}
	data, err := bfs.getFile(bfi)
	return HashFile{
		data: data,
		info: bfi,
	}, err
}

func (bfs *hashFileSystem) Delete(key string) error {
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
	if bfs.occupied == 0 {
		panic("[Delete Panic] Occupied is 0")
	}
	//update Occupied
	bfs.occupied -= bfi.Size_
	return nil
}

func (bfs *hashFileSystem) Opt(opt any) any {
	return nil
}

func (bfs *hashFileSystem) isExist(key string) bool {
	if key == "" {
		return false
	}
	_, err := bfs.getFileInfo(key)
	return err == nil
}

func (bfs *hashFileSystem) Cap() int64 {
	return bfs.capacity
}

// unit can be "B", "KB", "MB", "GB" or just leave it blank
func (bfs *hashFileSystem) Occupied(unit ...string) float64 {
	if len(unit) == 0 {
		return float64(bfs.occupied)
	}
	switch unit[0] {
	case "B":
		return float64(bfs.occupied)
	case "KB":
		return float64(bfs.occupied) / 1024
	case "MB":
		return float64(bfs.occupied) / 1024 / 1024
	case "GB":
		return float64(bfs.occupied) / 1024 / 1024 / 1024
	default:
		return float64(bfs.occupied)
	}
}

func (bfs *hashFileSystem) storeFile(key HashFileInfo, value []byte) error {
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

func (bfs *hashFileSystem) getFile(key HashFileInfo) ([]byte, error) {
	path := key.Path_
	if path == "" {
		return nil, errors.New("path is empty")
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

func (bfs *hashFileSystem) deleteFile(bfi HashFileInfo) error {
	fullPath := bfi.Path_ + "/" + bfi.FileName
	if fullPath == "" {
		return fmt.Errorf("path is empty")
	}
	err := os.Remove(fullPath)
	return err
}

func (bfs *hashFileSystem) getFileInfo(hashSum string) (HashFileInfo, error) {
	infoBytes, err := bfs.levelDB.Get([]byte(hashSum), nil)
	if err != nil {
		return HashFileInfo{}, err
	}
	var info HashFileInfo
	// pbfi := &pb.FileInfo{}
	// err = proto.Unmarshal(infoBytes, pbfi)
	err = json.Unmarshal(infoBytes, &info)
	return info, err
}

func (bfs *hashFileSystem) storeFileInfo(hashSum string, file HashFileInfo) error {
	if file.FileName == "" {
		return ErrFileInvalidName
	}
	// pbfi := &pb.FileInfo{}
	// HashFileInfoToPBFileInfo(file, pbfi)
	// res, err := proto.Marshal(pbfi)
	res, err := json.Marshal(file)
	if err != nil {
		return err
	}
	err = bfs.levelDB.Put([]byte(hashSum), res, nil)
	return err
}

func (bfs *hashFileSystem) deleteFileInfo(hashSum string) error {
	return bfs.levelDB.Delete([]byte(hashSum), nil)
}

func storeCapAndOccupied(levelDB *leveldb.DB, capacity, Occupied int64) error {
	if levelDB == nil {
		panic("levelDB is nil")
	}
	var capAndOccupied struct {
		Capacity int64 `json:"capacity"`
		Occupied int64 `json:"Occupied"`
	}
	capAndOccupied.Capacity = capacity
	capAndOccupied.Occupied = Occupied
	res, err := json.Marshal(capAndOccupied)
	if err != nil {
		return err
	}
	err = levelDB.Put([]byte("cap_and_Occupied"), res, nil)
	return err
}

func getCapAndOccupied(levelDB *leveldb.DB) (int64, int64, error) {
	if levelDB == nil {
		panic("levelDB is nil")
	}
	res, err := levelDB.Get([]byte("cap_and_Occupied"), nil)
	if err != nil {
		return 0, 0, err
	}
	var capAndOccupied struct {
		Capacity int64 `json:"capacity"`
		Occupied int64 `json:"Occupied"`
	}
	err = json.Unmarshal(res, &capAndOccupied)
	if err != nil {
		return 0, 0, err
	}
	return capAndOccupied.Capacity, capAndOccupied.Occupied, nil
}

func (bfs *hashFileSystem) Close() error {
	if bfs == nil {
		panic("hashFileSystem is nil")
	}
	log.Println("hashFileSystem Closing.")

	//save cap and ouppy
	if err := storeCapAndOccupied(bfs.levelDB, bfs.capacity, bfs.occupied); err != nil {
		log.Println("Save filesystem error:", err)
	}
	return bfs.levelDB.Close()
}

func (b HashFile) Data() []byte {
	return b.data
}

func (b HashFile) Stat() HashFileInfoI {
	return b.info
}

func NewFileInfo(fileName string, hashSum string, path string, size int64, modTime time.Time) HashFileInfo {
	bfi := HashFileInfo{
		FileName: fileName,
		Hash_:    hashSum,
		Path_:    path,
		Size_:    size,
		ModTime_: modTime,
	}
	return bfi
}

func (bfi HashFileInfo) Name() string {
	return bfi.FileName
}

func (bfi HashFileInfo) Path() string {
	return bfi.Path_
}

func (bfi HashFileInfo) Hash() string {
	return bfi.Hash_
}

func (bfi HashFileInfo) Size() int64 {
	return int64(bfi.Size_)
}

func (bfi HashFileInfo) ModTime() time.Time {
	return time.Time{}
}

func (bfi HashFileInfo) Mode() os.FileMode {
	return 0
}

func (bfi HashFileInfo) PeerInfo() peers.PeerInfo {
	return nil
}
