package service

import (
	"errors"
	"log"
	"strconv"
	"strings"
)

type fileRange struct {
	start, end int64
}

func parseRangeHeader(rangeHeader string, fileSize int64) (ranges []fileRange, err error) {
	const prefix = "bytes="
	if !strings.HasPrefix(rangeHeader, prefix) {
		return nil, errors.New("invalid range header")
	}

	rangeHeader = rangeHeader[len(prefix):]
	fileRanges := strings.Split(rangeHeader, ",")
	ranges = []fileRange{}

	for _, spec := range fileRanges {
		parts := strings.SplitN(spec, "-", 2)
		log.Printf("parts: %v\n", parts)
		if len(parts) != 2 {
			return nil, errors.New("invalid range header")
		}

		start, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return nil, err
		}
		var end int64
		if parts[1] == "" {
			end = fileSize - 1
		} else {
			end, err = strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return nil, err
			}

		}

		// 处理负数范围
		if start < 0 {
			start = fileSize + start
		}
		if end < 0 {
			end = fileSize + end
		}

		if start > end || end >= fileSize {
			return nil, errors.New("invalid range values")
		}

		ranges = append(ranges, fileRange{start, end})
	}

	return ranges, nil
}
