package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "devbrowser",
	Short:   "Launch an isolated Chrome session per git worktree",
	Version: version,
	Long: `devbrowser bridges git worktrees with isolated browser sessions.

Each worktree gets its own Chrome profile (cookies, sessions, storage)
so you can work on multiple features in parallel without interference.`,
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(stopCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, color.RedString("error:"), err)
		os.Exit(1)
	}
}
