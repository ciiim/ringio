package fs

import (
	"log"
)

type debug struct {
	debugLogger *log.Logger
	activeDebug bool
}

var dlog debug = debug{
	activeDebug: false,
}

func DebugOn() {
	dlog.activeDebug = true
	dlog.debugLogger = log.New(log.Writer(), "[DEBUG] ", log.Ltime|log.Lshortfile)
	dlog.debugLogger.SetFlags(log.Ltime | log.Lshortfile)
}

func IsDebug() bool {
	return dlog.activeDebug
}

func (dl *debug) debug(actionName string, format string, a ...any) {
	if dl.activeDebug {
		dl.debugLogger.Printf(actionName+": "+format, a...)
	}
}
