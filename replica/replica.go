// 存储服务多副本相关操作
// 实现副本存储、获取、删除等操作
// 通过node服务获取节点信息
package replica

import (
	"github.com/ciiim/cloudborad/node"
)

type ReplicaService = ReplicaServiceG[any]

func New(count int, ns *node.NodeServiceRO) *ReplicaService {
	return NewG[any](count, ns)
}
