package chash

import (
	"hash/crc64"
	"sort"
	"strconv"
	"sync"
)

type ConsistentHashFn func([]byte) uint64

type CHashItem interface {
	ID() int64
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

func (i *innerItem) ID() int64 {
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

	// hashFn hash函数
	hashFn ConsistentHashFn

	//replicas 虚拟节点个数
	replicas int

	hashRingMutex sync.RWMutex

	//hash ring 包含虚拟节点
	hashRing []int

	//node info map 包含虚拟节点
	hashMap map[int]CHashItem
}

func NewConsistentHash(replicas int, fn ConsistentHashFn) *ConsistentHash {
	m := &ConsistentHash{
		replicas:      replicas,
		hashFn:        fn,
		hashRingMutex: sync.RWMutex{},
		hashRing:      make([]int, 0),
		hashMap:       make(map[int]CHashItem),
	}

	if fn == nil {
		table := crc64.MakeTable(crc64.ISO)
		f := func(b []byte) uint64 {
			return crc64.Checksum(b, table)
		}
		m.hashFn = f
	}
	return m

}

func (c *ConsistentHash) Add(item CHashItem) {
	c.hashRingMutex.Lock()
	defer c.hashRingMutex.Unlock()

	//添加真实节点
	hashid := int(c.hashFn([]byte(strconv.FormatInt(item.ID(), 10))))

	//添加真实节点到hashMap
	c.hashMap[hashid] = warp2InnerItem(item, false)

	//添加真实节点到hashRing
	c.hashRing = append(c.hashRing, hashid)

	//添加虚拟节点
	for i := 0; i < c.replicas; i++ {
		hashid := int(c.hashFn([]byte(strconv.Itoa(i) + strconv.FormatInt(item.ID(), 10))))
		c.hashMap[hashid] = warp2InnerItem(item, true)
		c.hashRing = append(c.hashRing, hashid)
	}
	sort.Ints(c.hashRing)
}

func (c *ConsistentHash) Del(item CHashItem) {
	c.hashRingMutex.Lock()
	defer c.hashRingMutex.Unlock()

	//删除真实节点
	hash := int(c.hashFn([]byte(strconv.FormatInt(item.ID(), 10))))
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
		hash := int(c.hashFn([]byte(strconv.Itoa(i) + strconv.FormatInt(item.ID(), 10))))
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
}

func (c *ConsistentHash) Get(key []byte) CHashItem {

	c.hashRingMutex.RLock()
	defer c.hashRingMutex.RUnlock()

	if len(c.hashRing) == 0 {
		return nil
	}
	hash := int(c.hashFn(key))
	index := sort.Search(len(c.hashRing), func(i int) bool { return c.hashRing[i] >= hash })
	item := c.hashMap[c.hashRing[index%len(c.hashRing)]]
	realItem, ok := item.(*innerItem)
	if ok {
		return realItem.real

	}
	return nil
}

// 获取key最接近的节点以及后 n 个不相同的节点
// 如果节点数不足 n 个，则返回所有节点
func (c *ConsistentHash) GetN(key []byte, n int) []CHashItem {

	c.hashRingMutex.RLock()
	defer c.hashRingMutex.RUnlock()

	if n <= 0 {
		return nil
	}

	if len(c.hashRing) == 0 {
		return nil
	}

	if len(c.hashRing) <= n {
		items := make([]CHashItem, len(c.hashRing))
		for i := 0; i < len(c.hashRing); i++ {
			items[i] = c.hashMap[c.hashRing[i]].(*innerItem).real
		}
		return items
	}

	hash := int(c.hashFn(key))
	index := sort.Search(len(c.hashRing), func(i int) bool { return c.hashRing[i] >= hash })
	items := make([]CHashItem, n)

	items[0] = c.hashMap[c.hashRing[index%len(c.hashRing)]]

	for i := 0; i < n; {
		item := c.hashMap[c.hashRing[(index+i)%len(c.hashRing)]]
		if compare(item.(*innerItem), items[:i]...) {
			continue
		}
		items[i] = item.(*innerItem).real
		i++
	}
	items[0] = items[0].(*innerItem).real
	return items
}

func compare(a CHashItem, b ...CHashItem) bool {
	for _, v := range b {
		if a.Compare(v) {
			return true
		}
	}
	return false
}
