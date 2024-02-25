package replica

import (
	"errors"
	"slices"
)

var (
	ErrReplicaInfoNotFound = errors.New("replica info not found")

	ErrReplicaInfoKeyMismatch = errors.New("replica info key mismatch")

	ErrReplicaInfoCountMismatch = errors.New("replica info count mismatch")

	ErrReplicaInfoChecksumMismatch = errors.New("replica info checksum mismatch")

	ErrReplicaInfoAllNodesMismatch = errors.New("replica info all nodes mismatch")
)

/*
所有访问请求都会先访问主副本，如果找不到会将有效的第一个副本设置为主副本

默认情况下key对应的第一个副本为主副本

若一个远程用户请求在本地获取到此副本，视为原主副本节点宕机，将此副本提升为主副本,
*/
type ReplicaObjectInfoG[T CustomType] struct {

	// 副本数据Key
	Key []byte `json:"key"`

	Checksum []byte `json:"checksum"` // 校验和

	ExpectedReplicaCount int `json:"count"` // 期望副本数

	All []string `json:"all"` // 所有副本(包含主副本) nodeID

	Custom T `json:"-"` // 自定义数据，不存盘，用于存储一些临时数据
}

func NewReplicaObjectInfoG[T CustomType](key []byte, count int, all ...string) *ReplicaObjectInfoG[T] {
	return &ReplicaObjectInfoG[T]{
		Key:                  key,
		ExpectedReplicaCount: count,
		All:                  all,
	}
}

func (r *ReplicaObjectInfoG[T]) Set(a T) {
	r.Custom = a
}

func (r *ReplicaObjectInfoG[T]) Get() T {
	return r.Custom
}

func (r *ReplicaObjectInfoG[T]) Count() int {
	return r.ExpectedReplicaCount
}

// Sort 副本排序
// 在第一位的副本为主副本
func (r *ReplicaObjectInfoG[T]) Sort() {
	slices.Sort[[]string](r.All)
}
