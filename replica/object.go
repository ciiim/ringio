package replica

import "io"

// 副本对象
type ReplicaObject struct {
	io.ReadCloser
}

// 副本对象信息
type ReplicaObjectInfo struct {
	io.ReadWriteCloser
}

type Replica struct {
	object ReplicaObject
	info   ReplicaObjectInfo

	// 所有访问请求都会先访问主副本，如果找不到会将有效的第一个副本设置为主副本
	// 默认情况下key对应的第一个副本为主副本
	master bool

	others []string // 其他副本
}

func (r *Replica) SetMaster(master bool) {
	r.master = master
}

func (r *Replica) AddOther(others ...string) {
	r.others = append(r.others, others...)
}
