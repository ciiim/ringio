package replica

import "io"

type ReplicaReader struct {
	r     io.Reader
	count int // 副本数

	endpoints []string // 副本节点地址
}

func (r *ReplicaService) GetReader(reader io.Reader) *ReplicaReader {
	return &ReplicaReader{
		r:     reader,
		count: r.count,
	}
}
