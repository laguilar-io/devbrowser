//go:build darwin

package browser

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// WaitForClose blocks until all Chrome windows for the given PID are closed.
// On macOS, pressing X hides the window but keeps the process alive.
// We poll window count via osascript to detect the real close.
// Two consecutive zero-window readings are required to avoid false positives
// during page navigations or DevTools transitions.
func WaitForClose(pid int) {
	zeroCount := 0
	for {
		time.Sleep(1 * time.Second)

		// If process died, we're done
		if err := syscall.Kill(pid, 0); err != nil {
			return
		}

		// Check how many windows this Chrome process has open
		script := fmt.Sprintf(
			`tell application "System Events" to get count of windows of (first process whose unix id is %d)`,
			pid,
		)
		out, err := exec.Command("osascript", "-e", script).Output()
		if err != nil {
			// osascript failed (e.g. accessibility not granted) — fall back to process death
			zeroCount = 0
			continue
		}

		count, err := strconv.Atoi(strings.TrimSpace(string(out)))
		if err != nil {
			zeroCount = 0
			continue
		}

		if count == 0 {
			zeroCount++
			if zeroCount >= 2 {
				return // confirmed closed
			}
		} else {
			zeroCount = 0
		}
	}
}
