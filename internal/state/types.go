package state

import "time"

type Status string

const (
	StatusRunning Status = "running"
	StatusStale   Status = "stale"
)

type Entry struct {
	WorktreePath string    `json:"worktree_path"`
	Port         int       `json:"port"`
	ServerPID    int       `json:"server_pid"`
	ServerPGID   int       `json:"server_pgid"`
	BrowserPID   int       `json:"browser_pid"`
	Command      string    `json:"command"`
	StartedAt    time.Time `json:"started_at"`
}
