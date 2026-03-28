package worktree

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type Worktree struct {
	Path   string
	HEAD   string
	Branch string // empty for detached HEAD
	Name   string // basename of Path
}

// List returns all worktrees in the current git repo.
func List() ([]Worktree, error) {
	out, err := exec.Command("git", "worktree", "list", "--porcelain").Output()
	if err != nil {
		return nil, fmt.Errorf("not inside a git repository (or git not found)")
	}
	return parse(out), nil
}

// FindByName finds a worktree by name using a 3-step matching strategy:
// 1. Exact basename match
// 2. Branch name match (refs/heads/<name>)
// 3. Prefix match (unambiguous)
func FindByName(name string) (*Worktree, error) {
	wts, err := List()
	if err != nil {
		return nil, err
	}

	// Step 1: exact basename
	for i, wt := range wts {
		if wt.Name == name {
			return &wts[i], nil
		}
	}

	// Step 2: branch name
	for i, wt := range wts {
		branch := strings.TrimPrefix(wt.Branch, "refs/heads/")
		if branch == name {
			return &wts[i], nil
		}
	}

	// Step 3: prefix match — must be unambiguous
	var matches []Worktree
	for _, wt := range wts {
		if strings.HasPrefix(wt.Name, name) {
			matches = append(matches, wt)
		}
	}
	if len(matches) == 1 {
		return &matches[0], nil
	}
	if len(matches) > 1 {
		names := make([]string, len(matches))
		for i, m := range matches {
			names[i] = m.Name
		}
		return nil, fmt.Errorf("ambiguous worktree name %q — matches: %s", name, strings.Join(names, ", "))
	}

	return nil, fmt.Errorf("worktree %q not found", name)
}

func parse(data []byte) []Worktree {
	var result []Worktree
	var current Worktree
	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "worktree "):
			if current.Path != "" {
				result = append(result, current)
			}
			current = Worktree{
				Path: strings.TrimPrefix(line, "worktree "),
			}
			current.Name = filepath.Base(current.Path)
		case strings.HasPrefix(line, "HEAD "):
			current.HEAD = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			current.Branch = strings.TrimPrefix(line, "branch ")
		}
	}
	if current.Path != "" {
		result = append(result, current)
	}
	return result
}
