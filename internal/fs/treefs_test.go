package fs

import (
	"errors"
	"fmt"
	"testing"
)

const (
	trootPath = "./treefs"
)

func TestMkDir(t *testing.T) {
	tfs := newTreeFileSystem(trootPath)
	err := tfs.NewSpace("test", 1024*1024)
	if err != nil {
		if errors.Is(err, ErrSpaceExist) {
			t.Log(err)
		} else {
			t.Error(err)
			return
		}
	}
	space := tfs.GetSpace("test")
	for i := 0; i < 100; i++ {
		mkdir(space, "", fmt.Sprintf("bb%d", i))
	}
}

func mkdir(s *Space, base string, dir string) {
	err := s.makeDir(base, dir)
	if err != nil {
		if errors.Is(err, ErrSpaceExist) {
			return
		} else {
			panic(err)
		}
	}
}

func TestListDir(t *testing.T) {
	tfs := newTreeFileSystem(trootPath)
	space := tfs.GetSpace("test")
	defer space.Close(true)
	subs, err := space.getDir("", "aaa")
	if err != nil {
		t.Error(err)
	}
	for _, sub := range subs {
		t.Logf("isdir:%v name:%s", sub.IsDir(), sub.Name())
	}
}

func BenchmarkList(b *testing.B) {
	tfs := newTreeFileSystem(trootPath)
	space := tfs.GetSpace("test")
	defer space.Close(true)
	for i := 0; i < b.N; i++ {
		_, _ = space.getMetadata("aaa", "bbb.txt.meta")
	}
}

func TestStoreFile(t *testing.T) {
	tfs := newTreeFileSystem(trootPath)
	space := tfs.GetSpace("test")
	if space == nil {
		t.Error("space is nil")
		return
	}
	err := space.storeMetaData("aaa", "bbb.txt.meta", []byte("meta"))
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
	tfs := newTreeFileSystem(trootPath)
	space := tfs.GetSpace("test")
	defer space.Close(true)
	des, err := space.getDir("aaa", "bbb")
	if err != nil {
		t.Error(err)
		return
	}
	subs := DirEntryToSubInfo(des)
	for _, sub := range subs {
		t.Logf("isdir:%v name:%s", sub.IsDir, sub.Name)
	}
	err = space.deleteDir("aaa", "ccc")
	if err != nil {
		t.Error(err)
		return
	}
}
