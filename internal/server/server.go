package server

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/laguilar-io/devbrowser/internal/port"
)

type Server struct {
	cmd  *exec.Cmd
	PGID int
	PID  int
}

// Start launches the dev command inside dir with a new process group.
// port is injected via the PORT environment variable so it works with
// Next.js, Vite, CRA, and any other framework that respects PORT.
func Start(dir, command string, p int) (*Server, error) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = dir
	cmd.Stdin = nil
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Inject PORT — respected by Next.js, Vite, CRA, etc.
	cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", p))
	setSysProcAttr(cmd)

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start dev server: %w", err)
	}

	pgid := getPGID(cmd)
	return &Server{cmd: cmd, PGID: pgid, PID: cmd.Process.Pid}, nil
}

// Stop sends SIGTERM to the process group, waits up to 3s, then SIGKILL.
func (s *Server) Stop() {
	if s.PGID != 0 {
		killGroup(s.PGID)
	}
}

// Wait blocks until the dev server exits and returns its error.
func (s *Server) Wait() error {
	return s.cmd.Wait()
}

// WaitReady polls port until it accepts connections or timeout elapses.
func WaitReady(p int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	interval := 200 * time.Millisecond

	for time.Now().Before(deadline) {
		if port.IsOpen(p) {
			return nil
		}
		time.Sleep(interval)
		if interval < 2*time.Second {
			interval = interval * 3 / 2
		}
	}
	return fmt.Errorf("server did not become ready on port %d after %s", p, timeout)
}
