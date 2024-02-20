package replica

import (
	"errors"
	"io"
)

var (
	ErrNoAvailableReplica = errors.New("no available replica")
)

type ReplicaReader struct {

	// 副本
	key   []byte
	count int

	getReplica func(nodeid uint64, key []byte) (io.ReadCloser, error)

	readerCache io.ReadCloser

	// Read完成
	eof bool

	// 副本节点地址
	// 不包含主副本
	endpoints []uint64
}

func (r *ReplicaService) GetReader(key []byte) *ReplicaReader {
	return &ReplicaReader{
		key:   key,
		count: r.count,
	}
}

func (r *ReplicaReader) remoteReader() (io.ReadCloser, error) {
	// 读取副本节点
	for _, nodeid := range r.endpoints {
		reader, err := r.getReplica(nodeid, r.key)
		if err != nil {
			continue
		}
		return reader, nil
	}
	return nil, ErrNoAvailableReplica
}

func (r *ReplicaReader) Read(p []byte) (n int, err error) {
	if r.eof {
		return 0, io.EOF
	}

	// 读取缓存
	var reader io.ReadCloser
	if r.readerCache != nil {
		n, err = r.readerCache.Read(p)
	} else {
		// 读取副本节点
		reader, err = r.remoteReader()
		if err != nil {
			return 0, err
		}
		r.readerCache = reader
		n, err = reader.Read(p)
	}

	defer func() {
		if err == io.EOF {
			r.eof = true
		}
		if err != nil && r.readerCache != nil {
			r.readerCache.Close()
			r.readerCache = nil
		}
	}()
	return

}
