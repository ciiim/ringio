// implement hash chunk system
package hashchunk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ciiim/cloudborad/database"
	dlogger "github.com/ciiim/cloudborad/debug"
	"github.com/ciiim/cloudborad/storage/types"

	"github.com/syndtr/goleveldb/leveldb"
)

var (
	ErrEmptyKey = errors.New("key is empty")
)

type HashChunkSystem struct {
	rootPath string //相对路径 relative path

	config *Config

	capacity *types.SafeInt64
	occupied *types.SafeInt64

	chunkStatDBName string

	levelDB *leveldb.DB

	calcChunkStoragePathFn CalcChunkStoragePathFn

	HashFn Hash
}

type CalcChunkStoragePathFn = func(chunkStat *HashChunkInfo) string

type Hash func([]byte) []byte

// default calculate store path function
// format: year/month/day/chunkhash[0:3]/chunkhash[3:6]
var DefaultCalcStorePathFn = func(hci *HashChunkInfo) string {
	timePath := time.Time.Format(time.Now(), "2006/01/02")
	path := fmt.Sprintf("%s/%x/%x", timePath, hci.ChunkHash[0:3], hci.ChunkHash[3:6])
	return path
}

var _ IHashChunkSystem = (*HashChunkSystem)(nil)
var _ IHashChunkInfo = (*HashChunkInfo)(nil)

func NewHashChunkSystem(rootPath string, capacity int64, chunkSize int64, hashFn Hash, calcStoragePathFn CalcChunkStoragePathFn) *HashChunkSystem {
	if err := os.MkdirAll(rootPath, os.ModePerm); err != nil {
		panic("mkdir error:" + err.Error())
	}
	hashDBName := "chunkinfo"
	db, err := database.NewLevelDB(filepath.Join(rootPath + "/" + hashDBName))
	if err != nil {
		panic("leveldb init error:" + err.Error())
	}

	hcs := &HashChunkSystem{
		rootPath:               rootPath,
		capacity:               types.NewSafeInt64(),
		occupied:               types.NewSafeInt64(),
		chunkStatDBName:        hashDBName,
		levelDB:                db,
		calcChunkStoragePathFn: calcStoragePathFn,

		config: &Config{
			chunkMaxSize: chunkSize,
			hashFn:       hashFn,
		},
	}
	if calcStoragePathFn == nil {
		dlogger.Dlog.LogDebugf("[BFS]", "Use Default Calculate Function.")
		hcs.calcChunkStoragePathFn = DefaultCalcStorePathFn
	}

	cap, ouppy, err := hcs.getCapAndOccupied()

	if err != nil {
		log.Println("New HCS at", rootPath)
		_ = hcs.storeCapAndOccupied(capacity, 0)
		return hcs
	}
	log.Printf("Detect exist HCS at %s\n", rootPath)

	hcs.capacity.Store(cap)
	hcs.updateOccupied(ouppy)

	if capacity < cap {
		log.Println("[BFS] capacity is less than exist HCS, use exist HCS's capacity.")
	}
	if capacity > cap {
		log.Println("[BFS] capacity is more than exist HCS, use new capacity.")
		hcs.capacity.Store(capacity)
	}
	return hcs
}

func (hcs *HashChunkSystem) CreateChunk(key []byte, chunkName string) (io.WriteCloser, error) {
	if len(key) == 0 {
		return nil, ErrEmptyKey
	}
	// increase chunk counter
	// if chunk is exist, just increase counter
	// if chunk is not exist, create chunk info and store it
	_, err := hcs.increaseChunkCounter(key)
	if err == nil {
		return nil, fmt.Errorf("chunk is exist")
	}
	if err != nil && err != leveldb.ErrNotFound {
		return nil, err
	}

	hci := NewChunkInfo(chunkName, key, 0)
	hci.SetPath(filepath.Join(hcs.rootPath, hcs.calcChunkStoragePathFn(hci)))
	if err := os.MkdirAll(hci.ChunkPath, os.ModePerm); err != nil {
		return nil, err
	}
	if err := hcs.storeChunkInfo(key, hci); err != nil {
		return nil, err
	}
	return hcs.createChunkWriter(hci)
}

func (hcs *HashChunkSystem) StoreBytes(key []byte, chunkName string, value []byte) error {
	if len(key) == 0 {
		return ErrEmptyKey
	}
	if value == nil {
		return fmt.Errorf("value is nil")
	}

	// increase chunk counter
	// if chunk is exist, just increase counter
	// if chunk is not exist, create chunk info and store it
	_, err := hcs.increaseChunkCounter(key)
	if err == nil {
		return nil
	}
	if err != nil && err != leveldb.ErrNotFound {
		return err
	}

	valueLength := int64(len(value))

	//check capacity
	if err := hcs.CheckCapacity(valueLength); err != nil {
		return err
	}

	hci := NewChunkInfo(chunkName, key, valueLength)
	hci.SetPath(filepath.Join(hcs.rootPath, hcs.calcChunkStoragePathFn(hci)))

	// hci.Path = rootPath/<path>
	hci.ChunkPath = filepath.Join(hcs.rootPath, hcs.calcChunkStoragePathFn(hci))

	//make dir
	if err := os.MkdirAll(hci.ChunkPath, os.ModePerm); err != nil {
		return err
	}
	if err := hcs.storeChunkInfo(key, hci); err != nil {
		return err
	}
	if err := hcs.storeChunkBytes(hci, value); err != nil {
		return err
	}

	//update Occupied
	hcs.updateOccupied(hcs.occupied.Load() + hci.ChunkSize)

	return nil
}

func (hcs *HashChunkSystem) StoreReader(key []byte, chunkName string, v io.Reader) error {
	if len(key) == 0 {
		return ErrEmptyKey
	}
	if v == nil {
		return fmt.Errorf("value is nil")
	}

	// increase chunk counter
	// if chunk is exist, just increase counter
	// if chunk is not exist, create chunk info and store it
	_, err := hcs.increaseChunkCounter(key)
	if err == nil {
		return nil
	}

	if err != nil && err != leveldb.ErrNotFound {
		return err
	}

	//check capacity
	hci := NewChunkInfo(chunkName, key, 0)
	hci.SetPath(filepath.Join(hcs.rootPath, hcs.calcChunkStoragePathFn(hci)))
	if err := os.MkdirAll(hci.ChunkPath, os.ModePerm); err != nil {
		return err
	}
	if err := hcs.storeChunkInfo(key, hci); err != nil {
		return err
	}
	if err := hcs.storeChunkReader(hci, v); err != nil {
		return err
	}

	//update Occupied
	hcs.updateOccupied(hcs.occupied.Load() + hci.ChunkSize)

	return nil
}

func (hcs *HashChunkSystem) Get(key []byte) (*HashChunk, error) {
	if len(key) == 0 {
		return nil, ErrEmptyKey
	}
	hci, err := hcs.getChunkInfo(key)
	if err != nil {
		return nil, err
	}
	file, err := hcs.getChunk(hci)
	return &HashChunk{
		ReadCloser: file,
		info:       hci,
	}, err
}

func (hcs *HashChunkSystem) Delete(key []byte) error {
	if len(key) == 0 {
		return ErrEmptyKey
	}

	// decrease chunk counter
	nowCounter, err := hcs.decreaseChunkCounter(key)

	// still have reference
	if err == nil && nowCounter != 0 {
		return nil
	}

	if err != nil {
		return err
	}

	hci, err := hcs.getChunkInfo(key)
	if err != nil {
		return err
	}
	if hcs.occupied.Load()-hci.ChunkSize < 0 {
		return fmt.Errorf("[Delete Chunk Error] Occupied is 0")
	}
	if err := hcs.deleteChunkStat(key); err != nil {
		return err
	}
	if err := hcs.deleteChunk(hci); err != nil {
		return err
	}
	//update Occupied
	hcs.updateOccupied(hcs.occupied.Load() - hci.ChunkSize)
	return nil
}

func (hcs *HashChunkSystem) Opt(opt any) any {
	return nil
}

func (hcs *HashChunkSystem) isExist(key []byte) bool {
	_, err := hcs.getChunkInfo(key)
	return err == nil
}

func (hcs *HashChunkSystem) Cap() int64 {
	return hcs.capacity.Load()
}

// unit can be "B", "KB", "MB", "GB" or just leave it blank
func (hcs *HashChunkSystem) Occupied(unit ...string) float64 {
	if len(unit) == 0 {
		return float64(hcs.occupied.Load())
	}
	switch unit[0] {
	case "B":
		return float64(hcs.occupied.Load())
	case "KB":
		return float64(hcs.occupied.Load()) / 1024
	case "MB":
		return float64(hcs.occupied.Load()) / 1024 / 1024
	case "GB":
		return float64(hcs.occupied.Load()) / 1024 / 1024 / 1024
	default:
		return float64(hcs.occupied.Load())
	}
}

func (hcs *HashChunkSystem) createChunkWriter(hcStat *HashChunkInfo) (io.WriteCloser, error) {
	if hcs.calcChunkStoragePathFn == nil {
		return nil, fmt.Errorf("CalcChunkStoragePathFn is nil")
	}
	chunkFile, err := os.Create(filepath.Join(hcStat.ChunkPath, hcStat.ChunkName))
	if err != nil {
		return nil, fmt.Errorf("open file %s error: %s", hcStat.ChunkPath+"/"+hcStat.ChunkName, err)
	}
	chunkwc := warpHashChunkWriteCloser(chunkFile)
	return chunkwc, nil
}

func (hcs *HashChunkSystem) storeChunkBytes(hcStat *HashChunkInfo, value []byte) error {
	if hcs.calcChunkStoragePathFn == nil {
		return fmt.Errorf("CalcChunkStoragePathFn is nil")
	}
	file, err := os.Create(filepath.Join(hcStat.ChunkPath, hcStat.ChunkName))
	if err != nil {
		return fmt.Errorf("open file %s error: %s", hcStat.ChunkPath+"/"+hcStat.ChunkName, err)
	}
	defer file.Close()
	_, err = file.Write(value)
	return err
}

func (hcs *HashChunkSystem) storeChunkReader(key *HashChunkInfo, reader io.Reader) error {
	if hcs.calcChunkStoragePathFn == nil {
		return fmt.Errorf("CalcChunkStoragePathFn is nil")
	}
	file, err := os.Create(key.ChunkPath + "/" + key.ChunkName)
	if err != nil {
		return fmt.Errorf("open file %s error: %s", key.ChunkPath+"/"+key.ChunkName, err)
	}
	_, err = file.ReadFrom(reader)
	return err
}

func (hcs *HashChunkSystem) getChunk(hcStat *HashChunkInfo) (io.ReadCloser, error) {
	path := hcStat.ChunkPath
	if path == "" {
		return nil, errors.New("path is empty")
	}

	file, err := os.Open(filepath.Join(path, hcStat.ChunkName))
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (hcs *HashChunkSystem) deleteChunk(hci *HashChunkInfo) error {
	fullPath := filepath.Join(hci.ChunkPath, hci.ChunkName)
	err := os.Remove(fullPath)
	return err
}

func (hcs *HashChunkSystem) getChunkInfo(hashSum []byte) (*HashChunkInfo, error) {
	infoBytes, err := hcs.levelDB.Get(hashSum, nil)
	if err != nil {
		return nil, err
	}
	var info HashChunkInfo
	err = json.Unmarshal(infoBytes, &info)
	return &info, err
}

func (hcs *HashChunkSystem) storeChunkInfo(hashSum []byte, hci *HashChunkInfo) error {
	res, err := json.Marshal(hci)
	if err != nil {
		return err
	}
	err = hcs.levelDB.Put(hashSum, res, nil)
	return err
}

func (hcs *HashChunkSystem) deleteChunkStat(hashSum []byte) error {
	return hcs.levelDB.Delete(hashSum, nil)
}

func (hcs *HashChunkSystem) increaseChunkCounter(key []byte) (nowCounter int64, err error) {
	// get chunk info
	hci, err := hcs.getChunkInfo(key)
	if err != nil {
		return 0, err
	}
	hci.ChunkCount++

	// store chunk info
	return hci.ChunkCount, hcs.storeChunkInfo(key, hci)
}

func (hcs *HashChunkSystem) decreaseChunkCounter(key []byte) (nowCounter int64, err error) {
	// get chunk info
	hci, err := hcs.getChunkInfo(key)
	if err != nil {
		return 0, err
	}
	hci.ChunkCount--
	if hci.ChunkCount <= 0 {
		return 0, nil
	}

	// store chunk info
	return hci.ChunkCount, hcs.storeChunkInfo(key, hci)

}

func (hcs *HashChunkSystem) storeCapAndOccupied(capacity, occupied int64) error {
	var capAndOccupied struct {
		Capacity int64 `json:"capacity"`
		Occupied int64 `json:"occupied"`
	}
	capAndOccupied.Capacity = capacity
	capAndOccupied.Occupied = occupied
	res, err := json.Marshal(capAndOccupied)
	if err != nil {
		return err
	}
	err = hcs.levelDB.Put([]byte("cap_and_occupied"), res, nil)
	return err
}

func (hcs *HashChunkSystem) getCapAndOccupied() (int64, int64, error) {

	res, err := hcs.levelDB.Get([]byte("cap_and_occupied"), nil)
	if err != nil {
		return 0, 0, err
	}
	var capAndOccupied struct {
		Capacity int64 `json:"capacity"`
		Occupied int64 `json:"occupied"`
	}
	err = json.Unmarshal(res, &capAndOccupied)
	if err != nil {
		return 0, 0, err
	}
	return capAndOccupied.Capacity, capAndOccupied.Occupied, nil
}

func (hcs *HashChunkSystem) Close() error {
	if hcs == nil {
		panic("HashFileSystem is nil")
	}
	log.Println("HashFileSystem Closing.")

	//save cap and ouppy
	if err := hcs.storeCapAndOccupied(hcs.capacity.Load(), hcs.occupied.Load()); err != nil {
		log.Println("Save filesystem error:", err)
	}
	return hcs.levelDB.Close()
}

func (hcs *HashChunkSystem) updateOccupied(occupied int64) {
	hcs.occupied.Store(occupied)
}

func (hcs *HashChunkSystem) CheckCapacity(delta int64) error {
	if hcs.occupied.Load()+delta > hcs.capacity.Load() {
		return ErrFull
	}
	return nil
}

func (hcs *HashChunkSystem) Config() *Config {
	return hcs.config
}
