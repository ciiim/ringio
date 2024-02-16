package ringio

import (
	"github.com/ciiim/cloudborad/storage/hashchunk"
)

type DHashChunk struct {
	*hashchunk.HashChunk
}

func NewDHashChunk(h *hashchunk.HashChunk) *DHashChunk {
	return &DHashChunk{
		HashChunk: h,
	}
}
