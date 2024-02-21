package ringapi

import (
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"time"

	"github.com/ciiim/cloudborad/storage/tree"
	"github.com/ciiim/cloudborad/storage/types"
)

// 用户目录
type UserDir struct {
	FileNums int         `json:"file_nums"`
	Files    []*UserFile `json:"files"`
}

type UserFile struct {
	IsDir      bool   `json:"is_dir"`
	FileName   string `json:"file_name"`
	FileSize   int64  `json:"file_size"`
	ModTime    int64  `json:"mod_time"`
	CreateTime int64  `json:"create_time"`
}

func handleBaseDir(base string) string {
	if base == "" || base == "/" {
		base = tree.BASE_DIR
	}
	return base
}

func (r *RingAPI) NewSpace(space string, cap types.Byte) error {
	if cap < 0 {
		// FIXME: return error
		return nil
	}
	return r.ring.FrontSystem.NewSpace(space, cap)
}

func (r *RingAPI) MakeDir(space, base, name string) error {
	return r.ring.FrontSystem.MakeDir(space, base, name)
}

func (r *RingAPI) RenameDir(space, base, name, newName string) error {
	return r.ring.FrontSystem.RenameDir(space, base, name, newName)
}

func (r *RingAPI) DeleteDir(space, base, name string) error {
	return r.ring.FrontSystem.DeleteDir(space, base, name)
}

func (r *RingAPI) SpaceWithDir(space string, base, dir string) (*UserDir, error) {
	base = handleBaseDir(base)
	subInfo, err := r.ring.FrontSystem.GetDirSub(space, base, dir)
	if err != nil {
		return nil, err
	}
	userDir := &UserDir{}
	userDir.FileNums = len(subInfo)
	userDir.Files = make([]*UserFile, userDir.FileNums)
	for i, sub := range subInfo {
		userDir.Files[i] = &UserFile{
			IsDir:      sub.IsDir,
			FileName:   sub.Name,
			FileSize:   sub.Size,
			ModTime:    sub.ModTime.Unix(),
			CreateTime: sub.CreateTime.Unix(),
		}
	}
	return userDir, nil
}

func (r *RingAPI) PutFile(space, base, name string, fileHash []byte, fileSize int64, file multipart.File) error {
	base = handleBaseDir(base)
	chunkMaxSize := r.ring.StorageSystem.Config().ChunkMaxSize()
	if fileSize > chunkMaxSize {
		return r.putFileSplit(space, base, name, chunkMaxSize, fileHash, fileSize, file)
	}
	return r.putFile(space, base, name, fileHash, fileSize, file)
}

// 直接存储文件
func (r *RingAPI) putFile(space, base, name string, fileHash []byte, fileSize int64, file multipart.File) error {

	chunks := make([]*tree.FileChunk, 1)

	//创建FileChunk
	chunks[0] = tree.NewFileChunk(fileSize, fileHash)

	//存储元数据
	metadata := tree.NewMetaData(name, fileHash, time.Now(), chunks)
	metadataBytes, err := tree.MarshalMetaData(metadata)
	if err != nil {
		return err
	}
	if err = r.ring.FrontSystem.PutMetadata(space, base, name, fileHash, metadataBytes); err != nil {
		return err
	}

	//存储文件
	return r.ring.StorageSystem.StoreReader(fileHash, hex.EncodeToString(fileHash), file, nil)
}

// 分片存储文件
func (r *RingAPI) putFileSplit(space, base, name string, chunkSize int64, fileHash []byte, fileSize int64, file multipart.File) error {
	// 计算分片数量
	chunkNum := int(math.Ceil(float64(fileSize) / float64(chunkSize)))

	//创建分片Reader
	chunkReaders := make([]*io.SectionReader, chunkNum)
	chunks := make([]*tree.FileChunk, chunkNum)
	for i := 0; i < chunkNum-1; i++ {
		chunkReaders[i] = io.NewSectionReader(file, int64(i)*chunkSize, chunkSize)
	}

	//最后一个分片Reader
	chunkReaders[chunkNum-1] = io.NewSectionReader(file, int64(chunkNum-1)*chunkSize, fileSize-int64(chunkNum-1)*chunkSize)

	//计算每个分片的hash
	if err := func() error {
		chunkBuffer := r.chunkPool.Get()
		defer r.chunkPool.Put(chunkBuffer)
		for i, chunkReader := range chunkReaders {
			chunkBuffer.Reset()
			_, err := io.Copy(chunkBuffer, chunkReader)
			if err != nil {
				return err
			}
			chunks[i] = tree.NewFileChunk(
				chunkReader.Size(),
				chunkBuffer.Hash(r.ring.StorageSystem.Config().HashFn()),
			)
			if _, err := chunkReaders[i].Seek(0, io.SeekStart); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		return err
	}

	//存储元数据
	metadata := tree.NewMetaData(name, fileHash, time.Now(), chunks)
	metadataBytes, err := tree.MarshalMetaData(metadata)
	if err != nil {
		return err
	}

	if err = r.ring.FrontSystem.PutMetadata(space, base, name, fileHash, metadataBytes); err != nil {
		return err
	}

	//存储分片
	for i, chunkReader := range chunkReaders {
		if err := r.ring.StorageSystem.StoreReader(
			chunks[i].Hash,
			hex.EncodeToString(chunks[i].Hash),
			chunkReader,
			nil,
		); err != nil {
			return err
		}
	}

	return nil

}

type fileReader struct {
	io.ReadCloser
	FileSize int64
	FileName string
}

// 下载文件
func (r *RingAPI) GetFile(space, base, name string) (*fileReader, error) {
	base = handleBaseDir(base)
	metadataBytes, err := r.ring.FrontSystem.GetMetadata(space, base, name+".meta")
	if err != nil {
		fmt.Printf("GetMetadata: %v\n", err)
		return nil, err
	}
	var metadata tree.Metadata
	if err = tree.UnmarshalMetaData(metadataBytes, &metadata); err != nil {
		fmt.Printf("UnmarshalMetaData: %v\n", err)
		return nil, err
	}

	var (
		chunkClosers []io.Closer
		chunkReaders []io.Reader
		multiReader  io.Reader
	)

	for _, v := range metadata.Chunks {
		chunk, err := r.ring.StorageSystem.Get(v.Hash)
		if err != nil {
			fmt.Printf("GetFile: %v\n", err)
			return nil, err
		}
		chunkReaders = append(chunkReaders, chunk)
		chunkClosers = append(chunkClosers, chunk)
	}

	multiReader = io.MultiReader(chunkReaders...)

	return &fileReader{
		ReadCloser: multiReadCloser(multiReader, func() error {
			for _, v := range chunkClosers {
				if err := v.Close(); err != nil {
					return err
				}
			}
			return nil
		}),
		FileSize: metadata.Size,
		FileName: metadata.Filename,
	}, nil
}
