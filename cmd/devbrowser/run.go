package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/laguilar-io/devbrowser/internal/browser"
	"github.com/laguilar-io/devbrowser/internal/config"
	"github.com/laguilar-io/devbrowser/internal/port"
	"github.com/laguilar-io/devbrowser/internal/server"
	"github.com/laguilar-io/devbrowser/internal/state"
	"github.com/laguilar-io/devbrowser/internal/worktree"
)

var (
	flagPort    int
	flagCommand string
)

var runCmd = &cobra.Command{
	Use:   "run [worktree]",
	Short: "Start a dev server and open an isolated Chrome session",
	Long: `Start the dev server inside the given worktree (or the current directory),
wait for it to be ready, then open Chrome with an isolated profile.

When Chrome closes, the dev server is automatically stopped.`,
	Args:    cobra.MaximumNArgs(1),
	Aliases: []string{"r"},
	RunE:    runRun,
}

func init() {
	runCmd.Flags().IntVarP(&flagPort, "port", "p", 0, "Port override (default: auto-detect next free port from start_port)")
	runCmd.Flags().StringVarP(&flagCommand, "command", "c", "", "Dev server command override (default: config default_command)")

	// Make "devbrowser <worktree>" work as the default command
	rootCmd.RunE = runRun
	rootCmd.Args = cobra.MaximumNArgs(1)
	rootCmd.Flags().IntVarP(&flagPort, "port", "p", 0, "Port override")
	rootCmd.Flags().StringVarP(&flagCommand, "command", "c", "", "Dev server command override")
}

func runRun(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Determine working directory
	worktreeDir := "."
	worktreeName := ""

	if len(args) == 1 {
		wt, err := worktree.FindByName(args[0])
		if err != nil {
			return err
		}
		worktreeDir = wt.Path
		worktreeName = wt.Name
		fmt.Printf("  %s  %s\n", color.CyanString("worktree"), wt.Path)
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		worktreeDir = cwd
		worktreeName = filepath.Base(filepath.Dir(cwd)) + "__" + filepath.Base(cwd)
	}

	// Determine port
	p := flagPort
	if p == 0 {
		p, err = port.FindAvailable(cfg.StartPort)
		if err != nil {
			return err
		}
	}

	// Determine command
	devCmd := flagCommand
	if devCmd == "" {
		devCmd = cfg.DefaultCommand
	}
	fullCmd := fmt.Sprintf("%s -- -p %d", devCmd, p)

	// Check if already running
	existing, _ := state.Get(worktreeName)
	if existing != nil {
		return fmt.Errorf("already running on port %d (PID %d) — run `devbrowser stop %s` first",
			existing.Port, existing.ServerPID, worktreeName)
	}

	// Browser binary
	chromeBin, err := browser.Find(cfg.BrowserPath)
	if err != nil {
		return err
	}

	// Profile dir
	profilesDir := cfg.ProfilesDir
	if profilesDir == "" {
		profilesDir, err = config.ProfilesDir()
		if err != nil {
			return err
		}
	}
	profileDir := filepath.Join(profilesDir, worktreeName)

	url := fmt.Sprintf("http://localhost:%d", p)

	fmt.Printf("  %s  %s\n", color.CyanString("command "), fullCmd)
	fmt.Printf("  %s  %d\n", color.CyanString("port    "), p)
	fmt.Printf("  %s  %s\n", color.CyanString("profile "), profileDir)
	fmt.Printf("  %s  %s\n", color.CyanString("url     "), url)
	fmt.Println()

	// Start dev server
	srv, err := server.Start(worktreeDir, fullCmd)
	if err != nil {
		return err
	}

	// Save state
	entry := &state.Entry{
		WorktreePath: worktreeDir,
		Port:         p,
		ServerPID:    srv.PID,
		ServerPGID:   srv.PGID,
		Command:      fullCmd,
		StartedAt:    time.Now(),
	}
	_ = state.Add(worktreeName, entry)

	cleanup := func() {
		srv.Stop()
		_ = state.Remove(worktreeName)
	}

	// Wait for server to be ready
	fmt.Printf("⏳  Waiting for localhost:%d...\n", p)

	readyCh := make(chan error, 1)
	go func() {
		readyCh <- server.WaitReady(p, 90*time.Second)
	}()

	// Also watch for early server crash
	serverExitCh := make(chan error, 1)
	go func() {
		serverExitCh <- srv.Wait()
	}()

	select {
	case err := <-readyCh:
		if err != nil {
			cleanup()
			return err
		}
	case err := <-serverExitCh:
		cleanup()
		if err != nil {
			return fmt.Errorf("dev server exited unexpectedly: %w", err)
		}
		return fmt.Errorf("dev server exited unexpectedly")
	}

	fmt.Printf("✅  Server ready — opening Chrome...\n\n")

	// Launch browser
	browserCmd, err := browser.Launch(chromeBin, profileDir, url)
	if err != nil {
		cleanup()
		return err
	}
	entry.BrowserPID = browserCmd.Process.Pid
	_ = state.Add(worktreeName, entry)

	// Signal fan-in: browser close or OS signal
	done := make(chan string, 1)

	go func() {
		browserCmd.Wait()
		done <- "browser"
	}()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		done <- "signal"
	}()

	// Also watch server crash after browser is open
	go func() {
		serverExitCh <- srv.Wait()
	}()

	select {
	case reason := <-done:
		if reason == "browser" {
			fmt.Println("\n🔴  Chrome closed — stopping dev server...")
		} else {
			fmt.Println("\n🔴  Interrupted — stopping dev server...")
			_ = browserCmd.Process.Kill()
		}
	case <-serverExitCh:
		fmt.Println("\n🔴  Dev server exited — closing browser...")
		_ = browserCmd.Process.Kill()
	}

	cleanup()
	return nil
}
