package util

import (
	"fmt"
	"runtime"
)

func GetFuncName(skip int) string {
	pc := make([]uintptr, 1)
	runtime.Callers(skip, pc)
	f := runtime.FuncForPC(pc[0])
	_, lineNo := f.FileLine(pc[0])
	return fmt.Sprintf("%s:%d", f.Name(), lineNo)
}

func WarpWithDetail(err error) error {
	return fmt.Errorf("[%s] - %w", GetFuncName(3), err)
}
