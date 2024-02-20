package localtx

// 一个undo操作
// 记录在undo日志中
type undo struct {
	// 事务ID
	txID int64

	// 操作
	op op
}

// 一个操作
// func (tx *Tx) Op(opID int64, args ...any)
type op struct {

	// 操作id
	// 用于恢复时查找对应的操作函数
	OpID string
	// 操作参数
	Args []any
}
