// local transaction
package localtx

type TxState int

const (
	// TxDoing 事务进行中
	TxDoing TxState = iota

	// TxCommitting 事务提交中
	TxCommitting

	// TxCommitted 事务已提交
	TxCommitted

	// TxRollbacking 事务回滚中
	TxRollbacking

	// TxRollbacked 事务已回滚
	TxRollbacked
)

type Tx struct {
}
