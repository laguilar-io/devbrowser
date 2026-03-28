//go:build !windows

package server

import (
	"os/exec"
	"syscall"
	"time"
)

func setSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func getPGID(cmd *exec.Cmd) int {
	if cmd.Process == nil {
		return 0
	}
	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err != nil {
		return cmd.Process.Pid
	}
	return pgid
}

func killGroup(pgid int) {
	// SIGTERM first
	_ = syscall.Kill(-pgid, syscall.SIGTERM)
	// Give the process group 3 seconds to exit cleanly
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
