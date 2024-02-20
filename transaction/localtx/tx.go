// local transaction
package localtx

import (
	"os"
	"sync"
)

type TxState int

const (
	// TxDoing 事务进行中
	TxDoing TxState = iota

	// TxFailed 事务失败
	TxFailed

	// TxCommitting 事务提交中
	TxCommitting

	// TxCommitted 事务已提交
	TxCommitted

	// TxRollbacking 事务回滚中
	TxRollbacking

	// TxRollbacked 事务已回滚
	TxRollbacked
)

const (
	// OpSet 设置操作
	OpSet int8 = iota

	// OpUndo 撤销操作
	OpUndo
)

type OperationFn func(args ...any) error

type operation struct {

	// 操作类型
	OpType int8

	// 操作id
	OpID string

	// 操作参数
	Args []any

	// 关联函数
	Func OperationFn
}

type TxManager struct {
	// 事务日志路径
	undoLogPath string

	undoLogFile *os.File

	operations map[string]*operation
}

func (t *TxManager) OpenUndoLog() error {
	file, err := os.Open(t.undoLogPath)
	if err != nil {
		return err
	}

	t.undoLogFile = file
	return nil
}

type Tx struct {
	state TxState

	txLocker sync.Mutex

	ops []string

	undos []string

	manager *TxManager
}

func (tx *Tx) RegOp(opID string, opType int8, fn OperationFn, args ...any) {
	tx.manager.operations[opID] = &operation{
		OpType: opType,
		OpID:   opID,
		Args:   args,
		Func:   fn,
	}
	if opType == OpUndo {
		tx.undos = append(tx.undos, opID)
	} else {
		tx.ops = append(tx.ops, opID)
	}
}

func (tx *Tx) Commit() (TxState, error) {
	tx.txLocker.Lock()

	tx.state = TxCommitting
	for _, op := range tx.ops {
		op, ok := tx.manager.operations[op]
		if !ok {
			continue
		}
		if err := op.Func(op.Args); err != nil {
			tx.state = TxFailed
			return tx.state, err
		}
	}
	tx.state = TxCommitted
	tx.txLocker.Unlock()
	return tx.state, nil
}

func (tx *Tx) RollBack() (TxState, error) {
	defer tx.txLocker.Unlock()
	tx.state = TxRollbacking
	for _, op := range tx.undos {
		op, ok := tx.manager.operations[op]
		if !ok {
			continue
		}
		if err := op.Func(op.Args); err != nil {
			tx.state = TxFailed
			return tx.state, err
		}
	}
	tx.state = TxRollbacked
	return tx.state, nil
}
