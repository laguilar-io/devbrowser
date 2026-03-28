package state

import (
	"encoding/json"
	"os"

	"github.com/gofrs/flock"
	"github.com/laguilar-io/devbrowser/internal/config"
)

type State map[string]*Entry // key = worktree name

func load(path string) (State, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return State{}, nil
	}
	if err != nil {
		return State{}, err
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		// Corrupted — start fresh
		return State{}, nil
	}
	return s, nil
}

func save(path string, s State) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func withLock(fn func(path string, s State) error) error {
	stateFile, err := config.StateFile()
	if err != nil {
		return err
	}
	lockFile, err := config.StateLockFile()
	if err != nil {
		return err
	}

	fl := flock.New(lockFile)
	if err := fl.Lock(); err != nil {
		return err
	}
	defer fl.Unlock()

	s, err := load(stateFile)
	if err != nil {
		return err
	}
	return fn(stateFile, s)
}

func Add(name string, entry *Entry) error {
	return withLock(func(path string, s State) error {
		s[name] = entry
		return save(path, s)
	})
}

func Remove(name string) error {
	return withLock(func(path string, s State) error {
		delete(s, name)
		return save(path, s)
	})
}

func Get(name string) (*Entry, error) {
	stateFile, err := config.StateFile()
	if err != nil {
		return nil, err
	}
	s, err := load(stateFile)
	if err != nil {
		return nil, err
	}
	e, ok := s[name]
	if !ok {
		return nil, nil
	}
	return e, nil
}

func All() (State, error) {
	stateFile, err := config.StateFile()
	if err != nil {
		return nil, err
	}
	return load(stateFile)
}
