package chash

import (
	"hash/crc64"
	"slices"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"unsafe"
)

type ConsistentHashFn func([]byte) uint64

type CHashItem interface {
	ID() string
	Compare(o CHashItem) bool
}

// 用于实现向后查找真实节点
type innerItem struct {

	//virtual 是否是虚拟节点
	virtual bool

	//real 真实节点
	real CHashItem
}

func warp2InnerItem(real CHashItem, virtual bool) *innerItem {
	return &innerItem{
		virtual: virtual,
		real:    real,
	}
}

func (i *innerItem) ID() string {
	return i.real.ID()
}

func (i *innerItem) Compare(item CHashItem) bool {
	return i.real.ID() == item.ID()
}

func (i *innerItem) IsVirtual() bool {
	return i.virtual
}

// Consistent hash Map
type ConsistentHash struct {
	count int64

	// hashFn hash函数
	hashFn ConsistentHashFn

	//replicas 虚拟节点个数
	replicas int

	hashRingMutex sync.RWMutex

	//hash ring 包含虚拟节点
	hashRing []uint64

	//node info map 包含虚拟节点
	hashMap map[uint64]CHashItem

	//用于PickN的map
	findNMap map[string]CHashItem
}

var (
	DefaultHashFn = func(b []byte) uint64 {
		return crc64.Checksum(b, crc64.MakeTable(crc64.ISO))
	}
)

func NewConsistentHash(replicas int, fn ConsistentHashFn) *ConsistentHash {
	m := &ConsistentHash{
		replicas:      replicas,
		hashFn:        fn,
		hashRingMutex: sync.RWMutex{},
		hashRing:      make([]uint64, 0),
		hashMap:       make(map[uint64]CHashItem),
	}

	if fn == nil {
		m.hashFn = DefaultHashFn
	}
	return m

}

func (c *ConsistentHash) Len() int64 {
	return atomic.LoadInt64(&c.count)
}

// []byte READ ONLY
func string2Bytes(s string) (readOnly []byte) {
	sd := unsafe.StringData(s)
	return unsafe.Slice(sd, len(s))
}

func (c *ConsistentHash) GetByID(id string) CHashItem {
	c.hashRingMutex.RLock()
	defer c.hashRingMutex.RUnlock()
	item, ok := c.hashMap[c.hashFn(string2Bytes(id))]
	if !ok {
		return nil
	}
	return item.(*innerItem).real
}

func (c *ConsistentHash) Add(item CHashItem) {
	c.hashRingMutex.Lock()
	defer c.hashRingMutex.Unlock()

	//添加真实节点
	hashid := c.hashFn(string2Bytes(item.ID()))

	//添加真实节点到hashMap
	c.hashMap[hashid] = warp2InnerItem(item, false)

	//添加真实节点到hashRing
	c.hashRing = append(c.hashRing, hashid)

	//添加虚拟节点
	for i := 0; i < c.replicas; i++ {
		hashid := c.hashFn(string2Bytes(strconv.Itoa(i) + item.ID()))
		c.hashMap[hashid] = warp2InnerItem(item, true)
		c.hashRing = append(c.hashRing, hashid)
	}
	slices.Sort[[]uint64](c.hashRing)

	atomic.AddInt64(&c.count, 1)
}

func (c *ConsistentHash) Del(item CHashItem) {
	c.hashRingMutex.Lock()
	defer c.hashRingMutex.Unlock()

	//删除真实节点
	hash := c.hashFn(string2Bytes(item.ID()))
	i, found := sort.Find(len(c.hashRing), func(i int) int {
		return func() int {
			if c.hashRing[i] < hash {
				return -1
			} else if c.hashRing[i] == hash {
				return 0
			} else {
				return 1
			}
		}()
	})
	if !found {
		return
	}
	c.hashRing = append(c.hashRing[:i], c.hashRing[i+1:]...)
	delete(c.hashMap, hash)

	for i := 0; i < c.replicas; i++ {
		hash := c.hashFn(string2Bytes(strconv.Itoa(i) + item.ID()))
		delete(c.hashMap, hash)
		//删除虚拟节点
		i, found := sort.Find(len(c.hashRing), func(i int) int {
			return func() int {
				if c.hashRing[i] < hash {
					return -1
				} else if c.hashRing[i] == hash {
					return 0
				} else {
					return 1
				}
			}()
		})
		if !found {
			continue
		}
		c.hashRing = append(c.hashRing[:i], c.hashRing[i+1:]...)
	}

	atomic.AddInt64(&c.count, -1)
}

func (c *ConsistentHash) Get(key []byte) CHashItem {

	c.hashRingMutex.RLock()
	defer c.hashRingMutex.RUnlock()

	if len(c.hashRing) == 0 {
		return nil
	}
	hash := c.hashFn(key)

	index := sort.Search(len(c.hashRing), func(i int) bool { return c.hashRing[i] >= hash })
	item := c.hashMap[c.hashRing[index%len(c.hashRing)]]
	realItem, ok := item.(*innerItem)
	if ok {
		return realItem.real
	}
	return nil
}

// 获取key最接近的节点以及后 n 个不相同的节点
// 如果节点数不足 n 个，则返回所有不重复的节点
func (c *ConsistentHash) GetN(key []byte, n int) []CHashItem {
	if n <= 0 {
		return nil
	}

	c.hashRingMutex.RLock()
	defer c.hashRingMutex.RUnlock()

	if len(c.hashRing) == 0 {
		return nil
	}

	// 懒加载
	if c.findNMap == nil {
		c.findNMap = make(map[string]CHashItem)
	}

	// Find结束后清理findNMap
	defer func() {
		clear(c.findNMap)
	}()

	hash := c.hashFn(key)

	// 获取最接近的节点
	index := sort.Search(len(c.hashRing), func(i int) bool { return c.hashRing[i] >= hash })
	items := make([]CHashItem, n)

	items[0] = c.hashMap[c.hashRing[index%len(c.hashRing)]].(*innerItem).real

	// 剩余需要遍历的节点数
	remaining := len(c.hashRing) - 1

	// 获取后 n-1 个不相同的节点
	index++
	for count := 1; count < n && remaining > 0; func() {
		index++
		remaining--
	}() {
		item := c.hashMap[c.hashRing[index%len(c.hashRing)]]

		// 如果节点已经存在，则跳过
		if compare(item, c.findNMap) {
			continue
		}

		// 添加节点
		items[count] = item.(*innerItem).real
		c.findNMap[item.ID()] = item
		count++
	}
	return items
}

func compare(a CHashItem, m map[string]CHashItem) bool {
	_, ok := m[a.ID()]
	return ok
}
