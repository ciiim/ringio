package node

// ReadOnly service
type NodeServiceRO struct {
	self *Node

	//Read Only
	cMap *CMap
}

func (ns *NodeService) NodeServiceRO() *NodeServiceRO {
	if ns.ro == nil {
		ns.ro = &NodeServiceRO{
			cMap: ns.cMap,
			self: ns.self,
		}
	}
	return ns.ro
}

func (ns *NodeServiceRO) Self() *Node {
	return ns.self
}

func (ns *NodeServiceRO) Pick(key []byte) *Node {
	return ns.cMap.Get(key)
}

func (ns *NodeServiceRO) PickN(key []byte, n int) []*Node {
	return ns.cMap.GetN(key, n)
}

func (ns *NodeServiceRO) PickNext(key []byte) *Node {
	n := ns.cMap.GetN(key, 1)
	if len(n) == 2 {
		return n[1]
	}
	return n[0]
}
