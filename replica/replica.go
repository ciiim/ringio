// 存储服务多副本相关操作
// 实现副本存储、获取、删除等操作
// 通过node服务获取节点信息
package replica

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/ciiim/cloudborad/node"
)

var (
	ErrNoReplicaNode = errors.New("no replica node")
	ErrNoFunction    = errors.New("no function")
)

type ReplicaService struct {
	count int // 副本数
	ns    *node.NodeServiceRO

	putReplica func(nodeID string, reader io.Reader, info *ReplicaObjectInfo) error
	getReplica func(nodeID string, key []byte) (io.ReadCloser, *ReplicaObjectInfo, error)
	delReplica func(nodeID string, key []byte) error

	// 检查副本是否可用
	checkReplica func(nodeID string, info *ReplicaObjectInfo) error

	// 同步副本信息
	// 根据info中的节点信息同步到各个节点
	syncReplicaInfo func(info *ReplicaObjectInfo) error
}

func NewReplicaService(count int, ns *node.NodeServiceRO) *ReplicaService {
	return &ReplicaService{
		count: count,
		ns:    ns,
	}
}

// 注入副本存储相关函数实现
func (r *ReplicaService) SetFunctions(
	putReplica func(nodeID string, reader io.Reader, info *ReplicaObjectInfo) error,
	getReplica func(nodeID string, key []byte) (io.ReadCloser, *ReplicaObjectInfo, error),
	delReplica func(nodeID string, key []byte) error,
	checkReplica func(nodeID string, info *ReplicaObjectInfo) error,
	syncReplicaInfo func(info *ReplicaObjectInfo) error,
) {
	r.putReplica = putReplica
	r.getReplica = getReplica
	r.delReplica = delReplica
	r.checkReplica = checkReplica
	r.syncReplicaInfo = syncReplicaInfo
}

// 同步副本信息
// 必须在主副本节点上执行
// 建议放在后台队列执行
// 用于同步如chunk的引用计数等信息
func (r *ReplicaService) SyncReplicaInfo(info *ReplicaObjectInfo) error {
	if r.syncReplicaInfo == nil {
		return ErrNoFunction
	}
	return r.syncReplicaInfo(info)
}

// 检查副本是否可用
// 检查checksum是否一致，预防bit rot
// 若error不为nil，则副本不可用
func (r *ReplicaService) CheckReplica(nodeID string, info *ReplicaObjectInfo) error {
	if r.checkReplica == nil {
		return ErrNoFunction
	}
	return r.checkReplica(nodeID, info)
}

// 执行副本存储操作
// 必须在主副本节点上执行
// 建议放在后台队列执行
func (r *ReplicaService) PutReplica(key []byte, local io.ReadSeeker) (*ReplicaObjectInfo, error) {
	if r.putReplica == nil {
		return nil, ErrNoFunction
	}

	// 获取副本节点
	nodes := r.ns.PickN(key, r.count)
	if len(nodes) == 0 {
		return nil, ErrNoReplicaNode
	}

	// 不足副本所需的节点
	if len(nodes) < r.count {
		return nil, nil
	}

	nodeIDs := make([]string, 0, len(nodes))
	for _, node := range nodes {
		nodeIDs = append(nodeIDs, node.ID())
	}

	// 生成副本信息
	remoteInfo := NormalReplicaObjectInfo(key, r.count, nodeIDs...)

	// nodes[0] 为主副本
	// 从1开始为副本节点
	// 依次向副本节点写入数据
	for i := 1; i < len(nodes); i++ {
		if err := r.putReplica(nodes[i].ID(), local, remoteInfo); err != nil {
			return nil, err
		}

		// 重置reader
		if _, err := local.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
	}
	localInfo := MasterReplicaObjectInfo(key, r.count, nodeIDs...)
	return localInfo, nil
}

func (r *ReplicaService) GetReplica(key []byte) (reader io.ReadCloser, info *ReplicaObjectInfo, err error) {
	if r.getReplica == nil {
		return nil, nil, ErrNoFunction
	}

	nodes := r.ns.PickN(key, r.count)
	if len(nodes) == 0 {
		return nil, nil, ErrNoReplicaNode
	}

	// 依次从副本节点获取数据
	// 优先从主副本节点获取
	for _, node := range nodes {
		reader, info, err = r.getReplica(node.ID(), key)
		if err != nil {
			continue
		}
		return reader, info, nil
	}

	return nil, nil, ErrNoAvailableReplica
}

/*
从其它节点恢复副本

在新节点加入后，当有请求进来找不到对应的数据时，需要从其它节点恢复数据。

由于副本是往后备份的，当新节点加入后，只需要往后查找副本节点，
然后把多余的副本数据删除。

	e.g.
	n = 3

	key对应的节点为 2

	2(主) 3(副) 4(副) 5 6 ...

	新节点 1 加入后，该key对应的节点为 1

	1(主) 2(副) 3(副) 4(副) 5 6 ...
	^     ^           ^
	转主  转副         多余副本删除

	最终
	1(主) 2(副) 3(副) 4 5 6 ...
*/
func (r *ReplicaService) RecoverReplica(key []byte) (reader io.ReadCloser, info *ReplicaObjectInfo, err error) {
	if r.getReplica == nil {
		return nil, nil, ErrNoFunction
	}

	nodes := r.ns.PickN(key, r.count)
	if len(nodes) == 0 {
		return nil, nil, ErrNoReplicaNode
	}

	defer func() {
		if err == nil {
			// 移除冗余副本
			// 先进行副本健康检查，确定总副本数量，然后删除多余的副本
			go func() {
				_, _ = r.CheckAndAdjustReplica(info)
			}()
		}
	}()

	// 从副本节点恢复数据
	for i := 1; i < len(nodes); i++ {
		reader, info, err = r.getReplica(nodes[i].ID(), key)
		if err != nil {
			continue
		}
		return reader, info, nil
	}

	return nil, nil, ErrNoAvailableReplica
}

func (r *ReplicaService) DeleteReplica(info *ReplicaObjectInfo) error {

	if r.delReplica == nil {
		return ErrNoFunction
	}

	// 获取副本节点
	nodes := r.ns.PickN(info.Key, r.count)
	if nodes == nil {
		return ErrNoReplicaNode
	}

	// 依次删除副本节点数据
	for _, node := range nodes {
		if err := r.delReplica(node.ID(), info.Key); err != nil {
			return err
		}
	}
	return nil
}

func (r *ReplicaService) union(a, b []string) []string {
	m := make(map[string]struct{})
	for _, v := range a {
		m[v] = struct{}{}
	}
	for _, v := range b {
		m[v] = struct{}{}
	}
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

func (r *ReplicaService) diff(a, b []string) (aDiff []string, bDiff []string) {
	m := make(map[string]struct{})
	for _, v := range b {
		m[v] = struct{}{}
	}
	for _, v := range a {
		if _, ok := m[v]; !ok {
			aDiff = append(aDiff, v)
		}
	}
	clear(m)
	for _, v := range a {
		m[v] = struct{}{}
	}
	for _, v := range b {
		if _, ok := m[v]; !ok {
			bDiff = append(bDiff, v)
		}
	}
	return

}

type multiError struct {
	errs []error
}

func NewMultiError() *multiError {
	return &multiError{
		errs: make([]error, 0),
	}
}

func (m *multiError) Num() int {
	return len(m.errs)
}

func (m *multiError) Add(err error) {
	m.errs = append(m.errs, err)
}

func (m *multiError) Error() string {
	return fmt.Sprintf("multi error: %v", m.errs)
}

// 检查并调整副本，删除冗余的副本，增加新的副本
//
// 返回最后的在集群内的副本数量
//
// 较为耗时
// count = 3
func (r *ReplicaService) CheckAndAdjustReplica(oldInfo *ReplicaObjectInfo) (int64, error) {
	if r.checkReplica == nil || r.putReplica == nil || r.delReplica == nil {
		return 0, ErrNoFunction
	}

	count := atomic.Int64{}

	// 应该存放副本的节点
	newNodes := r.ns.PickN(oldInfo.Key, r.count)

	nodeIDs := make([]string, 0, len(newNodes))
	for _, node := range newNodes {
		nodeIDs = append(nodeIDs, node.ID())
	}

	var wg sync.WaitGroup

	wg.Add(2)

	multiErr := NewMultiError()

	go func() {

		defer wg.Done()

		actualKeep := int64(0)

		// 不需要调整的节点
		// 要检查副本是否存在
		keepNodes := r.union(oldInfo.All, nodeIDs)
		for _, nodeID := range keepNodes {
			if err := r.checkReplica(nodeID, oldInfo); err == nil {
				actualKeep++
			} else {
				multiErr.Add(err)
			}
		}

		count.Add(actualKeep)
	}()

	go func() {

		defer wg.Done()

		actualAddNodes := int64(0)

		// 新副本节点-冗余副本节点
		// 把副本传输到新副本节点，删除冗余副本节点的副本
		addNodes, redundantNodes := r.diff(oldInfo.All, nodeIDs)

		// 添加新副本
		for _, nodeID := range addNodes {
			reader, _, err := r.getReplica(oldInfo.All[0], oldInfo.Key)
			if err != nil {
				multiErr.Add(err)
				continue
			}
			if err := r.putReplica(nodeID, reader, oldInfo); err == nil {
				actualAddNodes++
			} else {
				multiErr.Add(err)
			}
			reader.Close()
		}

		// 删除冗余副本
		for _, nodeID := range redundantNodes {
			// 在原副本节点宕机的情况下，会无法执行删除操作，所以不视为错误
			if err := r.delReplica(nodeID, oldInfo.Key); err != nil && err != ErrNoReplicaNode {
				multiErr.Add(err)
			}
		}

		count.Add(actualAddNodes)
	}()

	wg.Wait()

	if multiErr.Num() > 0 {
		return count.Load(), multiErr
	}

	return count.Load(), nil

}
