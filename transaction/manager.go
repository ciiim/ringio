package transaction

import (
	"sync"
)

type TransactionID = string

type transactionState = int16

const (
	// 事务创建后的状态
	transactionStateInit transactionState = iota

	// 事务执行Try阶段
	transactionStateTrying

	// 事务执行Confirm阶段
	transactionStateConfirming

	// 事务执行Cancel阶段
	transactionStateCanceling

	// 事务执行完成
	transactionStateDone
)

type TransactionManager struct {
	tMu sync.RWMutex
	// 事务ID -> 事务
	// 从客户端传过来的数据的接收节点同时也是事务的发起和协调节点。
	// 出于简化考虑，这里不考虑持久化，只保证事务在发起节点不崩溃的情况下能够正常执行。
	transactions map[TransactionID]*transaction

	// 这里只对cancel进行重试，若try或confirm失败直接进入cancel阶段
	// cancel重试次数
	cancelRetry int

	// cancel重试次数
	cancelRetryInterval int

	// 超时时间
	retryTimeout int
}

func (tm *TransactionManager) NewTransaction(try, confirm, cancel func() error) *transaction {
	tm.tMu.Lock()
	defer tm.tMu.Unlock()
	//TODO: 生成事务ID
	return nil
}
