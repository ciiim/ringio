package types

import "sync/atomic"

type Byte = int64

type SafeInt64 = atomic.Int64

const (
	MB = Byte(8 << 17)
	GB = Byte(8 << 27)
)

func NewSafeInt64() *SafeInt64 {
	return &SafeInt64{}
}
