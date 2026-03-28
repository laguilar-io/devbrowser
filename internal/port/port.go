package port

import (
	"fmt"
	"net"
	"time"
)

// FindAvailable returns the first free TCP port starting from start.
// Uses Dial instead of Listen to correctly detect ports bound only to
// localhost (127.0.0.1) on macOS where SO_REUSEADDR can cause false negatives.
func FindAvailable(start int) (int, error) {
	for p := start; p <= 65535; p++ {
		if !IsOpen(p) {
			return p, nil
		}
	}
	return 0, fmt.Errorf("no available port found starting from %d", start)
}

// IsOpen returns true if a TCP connection can be established to the given port.
func IsOpen(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
