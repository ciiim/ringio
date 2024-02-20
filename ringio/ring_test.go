package ringio_test

import (
	"io"
	"os"
	"testing"
)

var (
	data = "hello world"
)

func TestTempFile(t *testing.T) {
	chunkTempFile, err := os.CreateTemp(os.TempDir(), "remote-chunk-")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		_, err := chunkTempFile.Write([]byte(data))
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err = chunkTempFile.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}

	chunkBuffer := make([]byte, 1024)
	for {
		n, err := chunkTempFile.Read(chunkBuffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(chunkBuffer[:n]))
	}
}

var (
	data1 = "hello world"
	data2 = "hello world2"
)

var (
	data3 = uint64(1234567890)
	data4 = uint64(1234567891)
)

func BenchmarkCompareString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if data1 == data2 {
			continue
		}
	}
}
func BenchmarkCompareUint64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if data3 == data4 {
			continue
		}
	}
}
