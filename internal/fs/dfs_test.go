package fs_test

import (
	"crypto/md5"
	"encoding/hex"
	"testing"
	"time"

	"github.com/ciiim/cloudborad/internal/fs"
)

const (
	rootPath = "./disk"

	//Byte
	capacity = 1024 * 1024 * 200

	replicas = 20
)

var (
	testDFileName = "hello.txt"
	testDFileData = "HelloHelloHelloHelloHelloHelloHelloHelloHelloHello"
)

var calcFileHash = func(data []byte) string {
	leng := len(data)
	md5Arr := md5.Sum(append(data, byte(leng)))
	md5Slice := md5Arr[:]
	return hex.EncodeToString(md5Slice)
}

func TestDFSPut(t *testing.T) {
	p := fs.NewDPeer("TestServer", "127.0.0.1", replicas, nil)
	dfs := fs.NewDFS(p, rootPath, capacity, nil)
	go dfs.Serve()
	time.Sleep(time.Second)
	hash := calcFileHash([]byte(testDFileData))
	err := dfs.Store(hash, testDFileName, []byte(testDFileData))
	if err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second * 5)
}

func TestDFSGet(t *testing.T) {
	p := fs.NewDPeer("TestServer", "127.0.0.1", replicas, nil)
	dfs := fs.NewDFS(p, rootPath, capacity, nil)
	go dfs.Serve()
	time.Sleep(time.Second)
	hash := calcFileHash([]byte(testDFileData))
	file, err := dfs.Get(hash)
	t.Logf("[Get Result]Info:%v,Error:%s", file.Stat(), err)
	time.Sleep(time.Second * 5)
}
