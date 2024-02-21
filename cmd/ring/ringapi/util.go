package ringapi

import "io"

type multiReadWithCloseFn struct {
	io.Reader
	close func() error
}

func (m *multiReadWithCloseFn) Close() error {
	return m.close()
}

func multiReadCloser(multiReader io.Reader, closeFn func() error) io.ReadCloser {
	return &multiReadWithCloseFn{
		Reader: multiReader,
		close:  closeFn,
	}
}
