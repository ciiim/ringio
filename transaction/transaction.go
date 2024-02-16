package transaction

type transaction struct {
	ID      TransactionID
	state   transactionState
	actions []action

	// 事务重试次数
	retry int
}
