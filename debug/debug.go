package dlogger

import (
	"log"
)

type debug struct {
	debugLogger *log.Logger
	printfSlot  func(actionName string, format string, a ...any)
	printlnSlot func(actionName string, a ...any)
	activeDebug bool
}

var Dlog debug = debug{
	printfSlot:  emptyFormatFn,
	printlnSlot: emptyLineFn,
	activeDebug: false,
}

var (
	debugFormatFn = func(actionName string, format string, a ...any) {
		Dlog.debugLogger.Printf(actionName+": "+format, a...)
	}
	emptyFormatFn = func(actionName string, format string, a ...any) {}

	debugLineFn = func(actionName string, a ...any) {
		Dlog.debugLogger.Println(actionName+": ", a)
	}
	emptyLineFn = func(actionName string, a ...any) {}
)

func DebugOn() {
	if Dlog.activeDebug {
		return
	}
	Dlog.activeDebug = true
	Dlog.printfSlot = debugFormatFn
	Dlog.printlnSlot = debugLineFn
	Dlog.debugLogger = log.New(log.Writer(), "[DEBUG] ", log.Ltime|log.Lshortfile)
}

func IsDebug() bool {
	return Dlog.activeDebug
}

func (dl *debug) LogDebugf(actionName string, format string, a ...any) {
	dl.printfSlot(actionName, format, a...)
}

func (dl *debug) LogDebug(actionName string, a ...any) {
	dl.printlnSlot(actionName, a...)
}
