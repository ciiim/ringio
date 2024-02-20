package replica

type ReplicaObjectInfo = ReplicaObjectInfoG[any]

func NewReplicaObjectInfo(key []byte, count int, all ...string) *ReplicaObjectInfo {
	return NewReplicaObjectInfoG[any](key, count, all...)
}
