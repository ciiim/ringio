package transaction

// 若error不为nil，表示该动作执行失败
type ActionFn func() error

type action struct {
	// 是否是远程动作
	remote bool

	// 重试次数
	timeout int

	tryFn     func() error
	confirmFn func() error
	cancelFn  func() error
}

func (a *action) Try() error {
	return a.tryFn()
}

func (a *action) Confirm() error {
	return a.confirmFn()
}

func (a *action) Cancel() error {
	return a.cancelFn()
}
