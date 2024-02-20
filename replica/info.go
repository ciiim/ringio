package replica

import "io"

// 副本对象
type ReplicaLocalObject interface {
	io.ReadSeekCloser
}

// 副本对象信息
type ReplicaObjectInfo struct {

	// 副本数据Key
	Key []byte `json:"key"`

	Checksum []byte `json:"checksum"` // 校验和

	/*
		所有访问请求都会先访问主副本，如果找不到会将有效的第一个副本设置为主副本

		默认情况下key对应的第一个副本为主副本

		若一个远程用户请求在本地获取到此副本，视为原主副本节点宕机，将此副本提升为主副本,

	*/
	Master bool `json:"master"` // 是否主副本

	ReplicaCount int `json:"count"` // 副本数

	All []string `json:"all"` // 所有副本(包含主副本) nodeID

	Custom map[string]string `json:"custom"` // 自定义数据
}

func MasterReplicaObjectInfo(key []byte, count int, all ...string) *ReplicaObjectInfo {
	return &ReplicaObjectInfo{
		Key:          key,
		Master:       true,
		ReplicaCount: count,
		All:          all,
	}
}

func NormalReplicaObjectInfo(key []byte, count int, all ...string) *ReplicaObjectInfo {
	return &ReplicaObjectInfo{
		Key:          key,
		Master:       false,
		ReplicaCount: count,
		All:          all,
	}
}

func (r *ReplicaObjectInfo) Set(key string, value string) {
	if r.Custom == nil {
		r.Custom = make(map[string]string)
	}
	r.Custom[key] = value
}

func (r *ReplicaObjectInfo) Get(key string) any {
	if r.Custom == nil {
		return nil
	}
	return r.Custom[key]
}

func (r *ReplicaObjectInfo) Delete(key string) {
	if r.Custom == nil {
		return
	}
	delete(r.Custom, key)
}

func (r *ReplicaObjectInfo) Count() int {
	return r.ReplicaCount
}

func (r *ReplicaObjectInfo) IsMaster() bool {
	return r.Master
}

func (r *ReplicaObjectInfo) ToMaster(me string) {
	r.Master = true
	for i, v := range r.All {
		if v == me {
			r.All[0], r.All[i] = r.All[i], r.All[0]
		}
	}
}

func (r *ReplicaObjectInfo) ToNormal() {
	r.Master = false
}
