package fs

import (
	"testing"
	"time"
)

func TestMetaData(t *testing.T) {
	blocksFileList := []Fileblock{
		newFileBlock("10.0.0.5", 1024, "hash1"),
		newFileBlock("10.0.0.7", 1024, "hash2"),
		newFileBlock("10.0.0.5", 1024, "hash3"),
	}
	meta := newMetaData("test", "testhash1", time.Now(), blocksFileList)
	data, err := marshalMetaData(&meta)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(string(data))
}
