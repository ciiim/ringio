package errmsg

import "fmt"

var (
	ErrDatabaseInternalError = fmt.Errorf("database internal error")
	ErrQueryUserFailed       = fmt.Errorf("query user failed")
	ErrInsertUserFailed      = fmt.Errorf("insert user failed")
)
