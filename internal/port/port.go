package port

import (
	"fmt"
	"net"
	"time"
)

// FindAvailable returns the first free TCP port starting from start.
func FindAvailable(start int) (int, error) {
	for p := start; p <= 65535; p++ {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", p))
		if err == nil {
			ln.Close()
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
