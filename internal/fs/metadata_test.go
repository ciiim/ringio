package fs

import (
	"testing"
	"time"
)

func TestMetaData(t *testing.T) {
	blocksFileList := []Fileblock{
		NewFileBlock("10.0.0.5", 1024, "hash1"),
		NewFileBlock("10.0.0.7", 1024, "hash2"),
		NewFileBlock("10.0.0.5", 1024, "hash3"),
	}
	meta := NewMetaData("test", "testhash1", time.Now(), blocksFileList)
	data, err := MarshalMetaData(&meta)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(string(data))
}
