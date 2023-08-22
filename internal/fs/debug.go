package fs

import (
	"log"
)

type debug struct {
	debugLogger *log.Logger
	activeDebug bool
}

var dlog debug

func DebugOn() {
	dlog.activeDebug = true
	dlog.debugLogger = log.Default()
	dlog.debugLogger.SetFlags(log.Ltime | log.Lshortfile)
}

func (dl *debug) debug(actionName string, format string, a ...any) {
	if dl.activeDebug {
		dl.debugLogger.Printf("[DEBUG] "+actionName+": "+format, a...)
	}
}
