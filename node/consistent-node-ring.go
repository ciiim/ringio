package node

import (
	"github.com/ciiim/cloudborad/node/chash"
)

type CMap struct {
	*chash.ConsistentHash
}

// create a new consistent hash map
func NewCMap(replicas int, fn chash.ConsistentHashFn) *CMap {
	return &CMap{
		chash.NewConsistentHash(replicas, fn),
	}
}

func (c *CMap) Add(node *Node) {
	c.ConsistentHash.Add(node)
}

func (c *CMap) Get(key []byte) *Node {
	node, ok := c.ConsistentHash.Get(key).(*Node)
	if !ok {
		return nil
	}
	return node
}

func (c *CMap) Del(node *Node) {
	c.ConsistentHash.Del(node)
}

func (c *CMap) GetN(key []byte, n int) []*Node {
	nodes := make([]*Node, 0, n)
	for _, node := range c.ConsistentHash.GetN(key, n) {
		nodes = append(nodes, node.(*Node))
	}
	return nodes
}
