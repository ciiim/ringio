package fs

import (
	"crypto/md5"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

var testFileName = "这是一个文件.txt"
var testData = strings.Repeat("文", 1024*5)
var testDataLen = len(testData)
var testCap int64 = 1024 * 1024 * 1024

func TestBasicFileSystem(t *testing.T) {
	f := newBasicFileSystem("./filestorage/block", testCap, nil)
	if f == nil {
		t.Error("newBasicFileSystem error")
	}
	defer f.Close()
}

func TestStore(t *testing.T) {
	f := newBasicFileSystem("./filestorage/block", 1024*10, nil)
	if f == nil {
		t.Error("newBasicFileSystem error")
	}
	md5 := fmt.Sprintf("%x", md5.Sum(append([]byte(testFileName), byte(testDataLen))))
	if err := f.Store(md5, testFileName, []byte(testData)); err != nil {
		t.Error(err)
		return
	}
	defer f.Close()
}

func TestGet(t *testing.T) {
	f := newBasicFileSystem("./filestorage/block", testCap, nil)
	if f == nil {
		t.Error("newBasicFileSystem error")
	}
	defer f.Close()
	md5 := fmt.Sprintf("%x", md5.Sum(append([]byte(testFileName), byte(testDataLen))))
	file, err := f.Get(md5[:])
	if err != nil {
		t.Error(err)
		return
	}
	info := file.Stat()
	t.Log("GetFile:", info.Path())
	t.Log("GetFileName:", info.Name())
	t.Log("GetFileHash:", info.Hash())
	t.Log("GetFileSize:", info.Size())
	t.Logf("occupy:%f", f.Occupy("KB"))
}

func TestOccupy(t *testing.T) {
	f := newBasicFileSystem("./filestorage/block", 1024*20, nil)
	if f == nil {
		t.Error("newBasicFileSystem error")
	}
	defer f.Close()
	fmt.Printf("f.Occupy(): %v\n", f.Occupy("MB"))
}

func TestBigFile(t *testing.T) {
	f := newBasicFileSystem("./filestorage/block", testCap, nil)
	if f == nil {
		t.Error("newBasicFileSystem error")
	}
	defer f.Close()
	//2 MB file
	filename0 := "test2MB0.txt"
	data0 := strings.Repeat("a23456d89A12s4567890123456789013", 163840)
	dataLen0 := len(data0)
	filename1 := "test2MB1.txt"
	data1 := strings.Repeat("a23456d89A1f34567890123456789013", 163840)
	dataLen1 := len(data0)
	filename2 := "test2MB2.txt"
	data2 := strings.Repeat("a23456d89A12345s7890123456789013", 163840)
	dataLen2 := len(data0)
	var wg sync.WaitGroup
	wg.Add(3)
	start := time.Now()
	go func() {
		md5 := fmt.Sprintf("%x", md5.Sum(append([]byte(data0), byte(dataLen0))))
		//md5Time := time.Since(start)
		if err := f.Store(md5, filename0, []byte(data0)); err != nil {
			t.Error(err)
			wg.Done()

			return
		}
		wg.Done()
	}()
	go func() {
		md5 := fmt.Sprintf("%x", md5.Sum(append([]byte(data1), byte(dataLen1))))
		//md5Time := time.Since(start)
		if err := f.Store(md5, filename1, []byte(data1)); err != nil {
			t.Error(err)
			wg.Done()

			return
		}
		wg.Done()

	}()
	go func() {
		md5 := fmt.Sprintf("%x", md5.Sum(append([]byte(data2), byte(dataLen2))))
		//md5Time := time.Since(start)
		if err := f.Store(md5, filename2, []byte(data2)); err != nil {
			t.Error(err)
			wg.Done()

			return
		}
		wg.Done()

	}()
	wg.Wait()
	delta := time.Since(start)
	t.Logf("use time:%v", delta)
}

func TestBigFileStore(t *testing.T) {
	f := newBasicFileSystem("./filestorage/block", testCap, nil)
	if f == nil {
		t.Error("newBasicFileSystem error")
	}
	defer f.Close()
	filename0 := "2test2MB.txt"
	data0 := strings.Repeat("c79456d99A12s41672901a34667d9711", 65536)
	dataLen0 := len(data0)
	md5 := fmt.Sprintf("%x", md5.Sum(append([]byte(data0), byte(dataLen0))))
	//md5Time := time.Since(start)
	start := time.Now()
	if err := f.Store(md5, filename0, []byte(data0)); err != nil {
		t.Error(err)
		return
	}
	delta := time.Since(start)
	t.Logf("use time:%v", delta)

}

func TestBigFileGet(t *testing.T) {
	f := newBasicFileSystem("./filestorage/block", testCap, nil)
	if f == nil {
		t.Error("newBasicFileSystem error")
	}
	defer f.Close()
	// filename0 := "2test2MB.txt"
	data0 := strings.Repeat("c79456d99A12s41672901a34667d9711", 65536)
	dataLen0 := len(data0)
	md5 := fmt.Sprintf("%x", md5.Sum(append([]byte(data0), byte(dataLen0))))
	//md5Time := time.Since(start)
	var wg sync.WaitGroup
	num := 4
	wg.Add(num)
	start := time.Now()
	for i := num; i > 0; i-- {
		go func() {
			for j := 0; j < 512/num; j++ {
				if _, err := f.Get(md5); err != nil {
					t.Error(err)
					return
				}
			}
			wg.Done()
		}()

	}
	wg.Wait()
	delta := time.Since(start)
	t.Logf("use time:%v", delta)

}
