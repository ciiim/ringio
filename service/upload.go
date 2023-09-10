package service

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"mime/multipart"
)

type PreUploadForm struct {
	Space       string `json:"space"`
	Base        string `json:"base"`
	FileName    string `json:"fileName"`
	FileHash    string `json:"fileHash"`
	TotalChunks int    `json:"totalChunks"`
}

type UploadForm struct {
	Space     string
	BaseDir   string
	StoreID   string // 用于识别分块文件
	ChunkHash string // 分块文件的hash

	File             *multipart.FileHeader // 文件
	ChunkNumber      int                   // 目前的分片序号
	ChunkSize        int64                 // 分片大小
	CurrentChunkSize int64                 // 当前分片大小
	TotalSize        int64                 // 文件总大小
	Filename         string                // 文件名
	RelativeDir      string                // 相对路径
	TotalChunks      int                   // 分片总数
}

func (s *Service) PreUploadFile(preUploadForm *PreUploadForm) (string, error) {
	id, err := s.fileServer.BeginStoreFile(preUploadForm.Space, preUploadForm.Base, preUploadForm.FileName, preUploadForm.FileHash, preUploadForm.TotalChunks)
	if err != nil {
		return "", err
	}
	return id, nil
}

/*
status:
-1:upload failed
0: begin upload
1: upload continue
2: upload success
*/
func (s *Service) UploadFile(uploadForm *UploadForm) (status int, err error) {
	if uploadForm == nil {
		return -1, fmt.Errorf("upload form is nil")
	}
	file, err := uploadForm.File.Open()
	if err != nil {
		return -1, err
	}
	defer file.Close()
	if status, err := s.fileServer.CheckUploadStatus(uploadForm.StoreID); status == -1 || err != nil {
		return -1, err
	}

	//get file data
	data := make([]byte, uploadForm.CurrentChunkSize)
	file.Read(data)

	chunkHash := md5.Sum(data)
	uploadForm.ChunkHash = hex.EncodeToString(chunkHash[:])
	log.Printf("upload chunk hash: %s", uploadForm.ChunkHash)

	if err := s.fileServer.StoreBlock(uploadForm.StoreID, uploadForm.ChunkNumber-1, uploadForm.ChunkHash, data); err != nil {
		return -1, err
	}
	return 1, nil
}

func (s *Service) UploadDone(storeID string) error {
	return s.fileServer.EndStoreFile(storeID)
}
