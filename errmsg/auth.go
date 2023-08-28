package errmsg

import "fmt"

var (
	ErrUserNotFound    = fmt.Errorf("user not found")
	ErrWrongPasswd     = fmt.Errorf("wrong passwd")
	ErrWeakPasswd      = fmt.Errorf("weak passwd")
	ErrWrongVerifyCode = fmt.Errorf("wrong verify code")
	ErrUserExist       = fmt.Errorf("user exist")
)
