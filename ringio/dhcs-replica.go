package ringio

import (
	"errors"
	"io"

	"github.com/ciiim/cloudborad/replica"
)

var (
	ErrSelfNode = errors.New("self node")
)

func (d *DHashChunkSystem) putReplica(
	nodeID string,
	reader io.Reader,
	info *replica.ReplicaObjectInfo,
) error {
	node := d.ns.GetByNodeID(nodeID)
	if node == nil {
		return replica.ErrNoReplicaNode
	}

	if node.Equal(d.ns.Self()) {
		return ErrSelfNode
	}

	d.remote.putReplica(node, reader, info)

	return nil
}
