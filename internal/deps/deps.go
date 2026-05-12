package deps

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureNodeModules checks if node_modules exists in worktreeDir.
// If missing, it symlinks from repoRoot/node_modules so no reinstall is needed.
// Returns a human-readable message describing what was done, or "" if nothing was needed.
func EnsureNodeModules(repoRoot, worktreeDir string) (string, error) {
	if repoRoot == "" || repoRoot == worktreeDir {
		return "", nil
	}

	target := filepath.Join(worktreeDir, "node_modules")
	if _, err := os.Lstat(target); err == nil {
		// Already exists (real dir or symlink)
		return "", nil
	}

	source := filepath.Join(repoRoot, "node_modules")
	if _, err := os.Stat(source); err != nil {
		// Root also has no node_modules — nothing to link
		return "", nil
	}

	if err := os.Symlink(source, target); err != nil {
		return "", fmt.Errorf("could not symlink node_modules: %w", err)
	}

	return fmt.Sprintf("node_modules → %s (symlinked)", shortenPath(source)), nil
}

func shortenPath(p string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return p
	}
	if len(p) > len(home) && p[:len(home)] == home {
		return "~" + p[len(home):]
	}
	return p
}
