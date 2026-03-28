//go:build !windows

package main

import (
	"syscall"
	"time"
)

func killGroupByPGID(pgid int) {
	if pgid <= 0 {
		return
	}
	_ = syscall.Kill(-pgid, syscall.SIGTERM)
	done := make(chan struct{})
	go func() {
		for {
			if err := syscall.Kill(-pgid, 0); err != nil {
				close(done)
				return
			}
			time.Sleep(200 * time.Millisecond)
		}
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
	}
}
