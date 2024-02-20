package chash_test

import (
	"hash/crc32"
	"hash/crc64"
	"strconv"
	"testing"
	"unsafe"
)

// []byte READ ONLY
func string2Bytes(s string) []byte {
	sd := unsafe.StringData(s)
	return unsafe.Slice(sd, len(s))
}

func TestCrc64(t *testing.T) {
	s := "hello"
	a := crc64.Checksum(string2Bytes(s), crc64.MakeTable(crc64.ISO))

	t.Log(a)
}

func BenchmarkCRC32(b *testing.B) {
	s := "hello"
	for i := 0; i < b.N; i++ {
		_ = crc32.ChecksumIEEE(string2Bytes(s))
	}
}

func BenchmarkCRC64(b *testing.B) {
	s := "hello"
	for i := 0; i < b.N; i++ {
		_ = crc64.Checksum(string2Bytes(s), crc64.MakeTable(crc64.ISO))
	}
}

func BenchmarkConvert0(b *testing.B) {
	s := "127.0.0.1:8800"

	for i := 0; i < b.N; i++ {
		_ = crc64.Checksum(string2Bytes(strconv.Itoa(i)+s), crc64.MakeTable(crc64.ISO))
	}
}

func BenchmarkConvert1(b *testing.B) {
	s := "127.0.0.1:8800"
	for i := 0; i < b.N; i++ {
		_ = crc64.Checksum([]byte(strconv.Itoa(i)+s), crc64.MakeTable(crc64.ISO))
	}
}
