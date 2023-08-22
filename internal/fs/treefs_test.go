package fs_test

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/ciiim/cloudborad/internal/fs"
)

const (
	trootPath = "./treefs"
)

func TestMkDir(t *testing.T) {
	tfs := fs.NewTreeFS(trootPath)
	space, err := tfs.NewSpace("test", 1024*1024)
	if err != nil {
		if errors.Is(err, fs.ErrSpaceExist) {
			t.Log(err)
		} else {
			t.Error(err)
			return
		}
	}
	for i := 0; i < 100; i++ {
		mkdir(space, fmt.Sprintf("bb%d", i))
	}
}

func mkdir(s *fs.Space, dir string) {
	err := s.MkDir(dir)
	if err != nil {
		if errors.Is(err, fs.ErrSpaceExist) {
			return
		} else {
			panic(err)
		}
	}
}

func TestListDir(t *testing.T) {
	tfs := fs.NewTreeFS(trootPath)
	space, _ := tfs.NewSpace("test", 1024*1024)
	defer space.Close(true)
	file, err := space.Get("")
	info := file.Stat()
	dirs := info.SubDir()
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("info.Name(): %v\n", info.Name())
	for _, dir := range dirs {
		size, err := space.GetSize(filepath.Join(info.Name(), dir.Name))
		if err != nil {
			t.Error(err)
			return
		}
		if dir.IsDir {
			t.Logf("[DIR ]%s --%d--", dir.Name, size)
		} else {
			t.Logf("[FILE]%s --%d--", dir.Name, size)
		}
	}
}

func BenchmarkList(b *testing.B) {
	tfs := fs.NewTreeFS(trootPath)
	space := tfs.GetSpace("test")
	defer space.Close(true)
	for i := 0; i < b.N; i++ {
		_, _ = space.Get("")
	}
}

func TestStoreFile(t *testing.T) {
	tfs := fs.NewTreeFS(trootPath)
	space := tfs.GetSpace("test")
	if space == nil {
		t.Error("space is nil")
		return
	}
	err := space.Store("aaa/bbb.txt", []byte("hello world"))
	if err != nil {
		t.Error(err)
		return
	}
	if err = space.Close(); err != nil {
		t.Error(err)
		return
	}

}

func TestDelDir(t *testing.T) {
	tfs := fs.NewTreeFS(trootPath)
	space := tfs.GetSpace("test")
	defer space.Close(true)
	file, err := space.Get("aaa")
	if err != nil {
		t.Error(err)
		return
	}
	info := file.Stat()
	dirs := info.SubDir()
	if err != nil {
		t.Error(err)
		return
	}
	for _, dir := range dirs {
		t.Logf("isdir:%v name:%s", dir.IsDir, dir.Name)
	}
	err = space.Delete("aaa/ccc")
	if err != nil {
		t.Error(err)
		return
	}
}
