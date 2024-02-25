package hashchunk

import (
	"errors"
)

var (
	ErrFull              = errors.New("chunk system is full")
	ErrChunkInfoNotFound = errors.New("chunk info not found")
	ErrChunkNotFound     = errors.New("chunk not found")
	ErrChunkExist        = errors.New("chunk already exist")
	ErrFileInvalidName   = errors.New("invalid file name")
	ErrNotDir            = errors.New("not a directory")
	ErrIsDir             = errors.New("is a directory")
	ErrInternal          = errors.New("internal error")
)
