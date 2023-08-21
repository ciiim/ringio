package fs

import "io/fs"

type Dir struct {
	DirName string
}

var _ DirEntry = (*Dir)(nil)

func (d Dir) Name() string {
	return d.DirName
}

func (d Dir) IsDir() bool {
	return true
}

func (d Dir) Type() fs.FileMode {
	return fs.ModeDir
}

func (d Dir) Info() (fs.FileInfo, error) {
	return nil, nil
}
