package ringapi_test

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestTmp(t *testing.T) {
	dir, err := os.MkdirTemp(os.TempDir(), "ringio-upload-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	t.Log(dir)
}

func TestSectionReader(t *testing.T) {
	s := "hello world"
	r := strings.NewReader(s)
	section := io.NewSectionReader(r, 0, 5)
	buf := make([]byte, 5)
	n, err := section.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(n, string(buf))
	_, _ = section.Seek(0, io.SeekStart)
	n, err = section.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(n, string(buf))
}
