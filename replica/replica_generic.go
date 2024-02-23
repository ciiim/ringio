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

var (
	ErrNoAvailableReplica = errors.New("no available replica")
)

type CustomType interface {
	any
}

type ReplicaServiceG[T CustomType] struct {
	count int // 副本数
	ns    *node.NodeServiceRO

	putReplica func(nodeID string, reader io.Reader, info *ReplicaObjectInfoG[T]) error
	getReplica func(nodeID string, key []byte) (io.ReadSeekCloser, *ReplicaObjectInfoG[T], error)
	delReplica func(nodeID string, key []byte) error

	// 检查副本信息是否可用和一致
	checkReplica func(nodeID string, info *ReplicaObjectInfoG[T]) error

	updateReplicaInfo func(nodeID string, info *ReplicaObjectInfoG[T]) error
}

func NewG[T CustomType](count int, ns *node.NodeServiceRO) *ReplicaServiceG[T] {
	return &ReplicaServiceG[T]{
		count: count,
		ns:    ns,
	}
}

/*
注入副本存储相关函数实现

需要分别实现
  - 副本存储
  - 副本获取
  - 副本删除
  - 副本检查
  - 副本信息同步
*/
func (r *ReplicaServiceG[T]) SetFunctions(
	putReplica func(nodeID string, reader io.Reader, info *ReplicaObjectInfoG[T]) error,
	getReplica func(nodeID string, key []byte) (io.ReadSeekCloser, *ReplicaObjectInfoG[T], error),
	delReplica func(nodeID string, key []byte) error,
	checkReplica func(nodeID string, info *ReplicaObjectInfoG[T]) error,
	updateReplicaInfo func(nodeID string, info *ReplicaObjectInfoG[T]) error,
) {
	r.putReplica = putReplica
	r.getReplica = getReplica
	r.delReplica = delReplica
	r.checkReplica = checkReplica
	r.updateReplicaInfo = updateReplicaInfo
}

// 同步副本信息
// 必须在主副本节点上执行
// 建议放在后台队列执行
// 用于同步如chunk的引用计数等信息
func (r *ReplicaServiceG[T]) UpdateReplicaInfo(nodeID string, info *ReplicaObjectInfoG[T]) error {
	if r.updateReplicaInfo == nil {
		return ErrNoFunction
	}
	return r.updateReplicaInfo(nodeID, info)
}

// 检查副本是否可用
// 检查checksum是否一致，预防bit rot
// 若error不为nil，则副本不可用
func (r *ReplicaServiceG[T]) CheckReplica(nodeID string, info *ReplicaObjectInfoG[T]) error {
	if r.checkReplica == nil {
		return ErrNoFunction
	}
	return r.checkReplica(nodeID, info)
}

// 执行副本存储操作
// 必须在主副本节点上执行
// 建议放在后台队列执行
func (r *ReplicaServiceG[T]) PutReplica(key []byte, local io.ReadSeeker) (*ReplicaObjectInfoG[T], error) {
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
	remoteInfo := NewReplicaObjectInfoG[T](key, r.count, nodeIDs...)

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
	localInfo := NewReplicaObjectInfoG[T](key, r.count, nodeIDs...)
	return localInfo, nil
}

func (r *ReplicaServiceG[T]) GetReplica(key []byte) (reader io.ReadSeekCloser, info *ReplicaObjectInfoG[T], err error) {
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

在新节点加入后，当有请求进来找不到对应的数据时，需要从其它节点恢复数据，
由于副本是往后备份的，当新节点加入后，只需要往后查找副本节点，然后把多余的副本数据删除。

当主副本节点宕机后，后一个节点成为该副本的主节点，其他节点成为副本节点，
由于副本是往后备份的，只需要把后一个节点的副本提升为主副本，然后添加缺少的副本节点。

	e.g. 新节点加入
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
func (r *ReplicaServiceG[T]) RecoverReplica(key []byte) (reader io.ReadSeekCloser, info *ReplicaObjectInfoG[T], err error) {
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

func (r *ReplicaServiceG[T]) DeleteReplica(info *ReplicaObjectInfoG[T]) error {

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

func (r *ReplicaServiceG[T]) union(a, b []string) []string {
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

func (r *ReplicaServiceG[T]) diff(a, b []string) (aDiff []string, bDiff []string) {
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

/*
检查并调整副本，删除冗余的副本，增加新的副本

返回最后的在集群内的副本数量

较为耗时

会修改传入的副本信息

	修改Info
	- All
	- ReplicaCount
	- Master
*/
func (r *ReplicaServiceG[T]) CheckAndAdjustReplica(info *ReplicaObjectInfoG[T]) (int64, error) {
	if r.checkReplica == nil || r.putReplica == nil || r.delReplica == nil {
		return 0, ErrNoFunction
	}

	var (
		count = atomic.Int64{}

		newInfo = *info

		oldNodeIDs = info.All
	)

	pn := r.ns.PickN(info.Key, r.count)
	newInfoNodeIDs := make([]string, 0, len(pn))
	for _, node := range pn {
		newInfoNodeIDs = append(newInfoNodeIDs, node.ID())
	}

	//更新副本节点
	newInfo.All = newInfoNodeIDs
	newInfo.Sort()

	var wg sync.WaitGroup

	wg.Add(2)

	multiErr := NewMultiError()

	// 新副本节点-冗余副本节点
	// 把副本传输到新副本节点，删除冗余副本节点的副本
	addNodes, redundantNodes := r.diff(info.All, newInfoNodeIDs)

	go func() {

		defer wg.Done()

		actualDeleteNodes := int64(0)

		// 删除冗余副本
		for _, nodeID := range redundantNodes {
			// 在原副本节点宕机的情况下，会无法执行删除操作，所以不视为错误
			if err := r.delReplica(nodeID, info.Key); err != nil && err != ErrNoReplicaNode {
				multiErr.Add(err)
			} else {
				actualDeleteNodes++
			}
		}

		count.Add(-actualDeleteNodes)

	}()

	go func() {

		defer wg.Done()

		actualAddNodes := int64(0)

		var (
			rsc io.ReadSeekCloser
			err error
		)

		//从旧节点中获取副本
		for _, node := range info.All {
			rsc, _, err = r.getReplica(node, info.Key)
			if err != nil {
				continue
			} else {
				break
			}
		}
		if rsc == nil {
			multiErr.Add(ErrNoAvailableReplica)
			return
		}
		defer rsc.Close()

		// 添加新副本
		for _, nodeID := range addNodes {
			if err := r.putReplica(nodeID, rsc, info); err == nil {
				actualAddNodes++
			} else {
				multiErr.Add(err)
			}

			// 重置reader
			if _, err = rsc.Seek(0, io.SeekStart); err != nil {
				multiErr.Add(err)
				return
			}
		}
		count.Add(actualAddNodes)
	}()

	wg.Wait()

	// 不需要调整的节点
	// 要检查副本是否存在
	actualKeep := int64(0)
	keepNodes := r.union(oldNodeIDs, newInfoNodeIDs)

	for _, nodeID := range keepNodes {
		if err := r.updateReplicaInfo(nodeID, &newInfo); err == nil {
			actualKeep++
		} else {
			multiErr.Add(err)
		}
	}
	count.Add(actualKeep)

	if multiErr.Num() > 0 {
		return count.Load(), multiErr
	}

	return count.Load(), nil

}

type multiError struct {
	mu   sync.Mutex
	errs []error
}

func NewMultiError() *multiError {
	return &multiError{
		errs: make([]error, 0),
	}
}

func (m *multiError) Num() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.errs)
}

func (m *multiError) Add(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errs = append(m.errs, err)
}

func (m *multiError) Error() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return fmt.Sprintf("multi error: %v", m.errs)
}
