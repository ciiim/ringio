package dlogger

import (
	"log"
)

type debug struct {
	debugLogger *log.Logger
	activeDebug bool
}

var Dlog debug = debug{
	activeDebug: false,
}

func DebugOn() {
	Dlog.activeDebug = true
	Dlog.debugLogger = log.New(log.Writer(), "[DEBUG] ", log.Ltime|log.Lshortfile)
	Dlog.debugLogger.SetFlags(log.Ltime | log.Lshortfile)
}

func IsDebug() bool {
	return Dlog.activeDebug
}

func (dl *debug) LogDebugf(actionName string, format string, a ...any) {
	if dl.activeDebug {
		dl.debugLogger.Printf(actionName+": "+format, a...)
	}
}
