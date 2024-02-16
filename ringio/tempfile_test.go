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
