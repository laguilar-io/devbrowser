//go:build windows

package main

import (
	"fmt"
	"os/exec"
)

func killGroupByPGID(pid int) {
	exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", pid)).Run()
}
