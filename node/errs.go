package node

import "errors"

var (
	ErrNodeNotFound = errors.New("node not found")
	ErrSelfNode     = errors.New("self node")
)
