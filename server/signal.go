//go:build linux
// +build linux

package server

import (
	"os"
	"os/signal"
	"syscall"
)

func WaitSIGINT(stopChan chan struct{}) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT)
	<-sig
	stopChan <- struct{}{}
}
