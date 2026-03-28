package main

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/laguilar-io/devbrowser/internal/state"
)

var flagStopAll bool

var stopCmd = &cobra.Command{
	Use:   "stop [worktree]",
	Short: "Stop a running dev server session",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		all, err := state.All()
		if err != nil {
			return err
		}

		if flagStopAll {
			if len(all) == 0 {
				fmt.Println("No active sessions.")
				return nil
			}
			for name, entry := range all {
				stopEntry(name, entry)
			}
			return nil
		}

		if len(args) == 0 {
			return fmt.Errorf("specify a worktree name or use --all")
		}

		name := args[0]
		entry, ok := all[name]
		if !ok {
			return fmt.Errorf("no session found for %q", name)
		}
		stopEntry(name, entry)
		return nil
	},
}

func init() {
	stopCmd.Flags().BoolVar(&flagStopAll, "all", false, "Stop all running sessions")
}

func stopEntry(name string, entry *state.Entry) {
	fmt.Printf("Stopping %s (port %d, PGID %d)...", color.CyanString(name), entry.Port, entry.ServerPGID)
	killGroupByPGID(entry.ServerPGID)
	_ = state.Remove(name)
	fmt.Println(" done")
}
