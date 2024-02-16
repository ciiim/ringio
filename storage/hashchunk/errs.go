package hashchunk

import (
	"errors"
)

var (
	ErrFull            = errors.New("file system is full")
	ErrFileNotFound    = errors.New("file not found")
	ErrFileExist       = errors.New("file or dir already exist")
	ErrFileInvalidName = errors.New("invalid file name")
	ErrNotDir          = errors.New("not a directory")
	ErrIsDir           = errors.New("is a directory")
	ErrInternal        = errors.New("internal error")
)
