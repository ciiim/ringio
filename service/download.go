package service

func (s *Service) PreDownloadFile(space, base, filename string) (downloadID string, fileSize int64, blockNum int, err error) {
	return s.fileServer.BeginDownloadFile(space, base, filename)
}

func (s *Service) DownloadChunk(downloadID string, chunkIndex int) ([]byte, error) {
	return s.fileServer.GetBlock(downloadID, chunkIndex)
}

func (s *Service) GetSizeByDownloadID(downloadID string) int64 {
	return s.fileServer.DownloadTaskInfo(downloadID).FileSize
}

func (s *Service) DownloadChunks(downloadID string, dataRange string) (chunk []byte, start int64, end int64, totalSize int64, err error) {
	totalSize = s.GetSizeByDownloadID(downloadID)
	ranges, err := parseRangeHeader(dataRange, totalSize)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	chunk, err = s.fileServer.GetBlockByRange(downloadID, ranges[0].start, ranges[0].end)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	s.fileServer.EndDownloadFile(downloadID)
	return chunk, ranges[0].start, ranges[len(ranges)-1].end, totalSize, nil
}

func (s *Service) DownloadDone(downloadID string) error {
	return s.fileServer.EndDownloadFile(downloadID)
}
