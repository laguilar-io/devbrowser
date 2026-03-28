package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/laguilar-io/devbrowser/internal/browser"
	"github.com/laguilar-io/devbrowser/internal/config"
	"github.com/laguilar-io/devbrowser/internal/envfiles"
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
	Use:     "run [worktree]",
	Short:   "Start a dev server and open an isolated Chrome session",
	Args:    cobra.MaximumNArgs(1),
	Aliases: []string{"r"},
	RunE:    runRun,
}

func init() {
	runCmd.Flags().IntVarP(&flagPort, "port", "p", 0, "Port override (default: auto-detect)")
	runCmd.Flags().StringVarP(&flagCommand, "command", "c", "", "Dev server command override")

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

	worktreeDir := "."
	worktreeName := ""
	repoRoot := ""

	if len(args) == 1 {
		wt, err := worktree.FindByName(args[0])
		if err != nil {
			return err
		}
		worktreeDir = wt.Path
		projectName := filepath.Base(filepath.Dir(filepath.Dir(wt.Path)))
		worktreeName = projectName + "__" + wt.Name
		// repo root = main worktree (first entry in git worktree list)
		wts, _ := worktree.List()
		if len(wts) > 0 {
			repoRoot = wts[0].Path
		}
		fmt.Printf("  %s  %s\n", color.CyanString("worktree"), wt.Path)
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		worktreeDir = cwd
		worktreeName = filepath.Base(filepath.Dir(cwd)) + "__" + filepath.Base(cwd)
		repoRoot = cwd
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

	fmt.Printf("  %s  %s\n", color.CyanString("command "), devCmd)
	fmt.Printf("  %s  %d\n", color.CyanString("port    "), p)
	fmt.Printf("  %s  %s\n", color.CyanString("profile "), profileDir)
	fmt.Printf("  %s  %s\n", color.CyanString("url     "), url)
	fmt.Println()

	// Copy .env*.local files from repo root to worktree
	var envResult *envfiles.CopyResult
	if repoRoot != "" && repoRoot != worktreeDir {
		envResult, err = envfiles.CopyToWorktree(repoRoot, worktreeDir)
		if err != nil {
			fmt.Printf("⚠️  Could not copy env files: %v\n", err)
		} else if len(envResult.Copied) > 0 {
			for _, f := range envResult.Copied {
				fmt.Printf("  %s  %s\n", color.CyanString("env     "), filepath.Base(f))
			}
			fmt.Println()
		}
	}

	// Start dev server
	srv, err := server.Start(worktreeDir, devCmd, p)
	if err != nil {
		return err
	}

	entry := &state.Entry{
		WorktreePath: worktreeDir,
		Port:         p,
		ServerPID:    srv.PID,
		ServerPGID:   srv.PGID,
		Command:      devCmd,
		StartedAt:    time.Now(),
	}
	_ = state.Add(worktreeName, entry)

	cleanup := func() {
		srv.Stop()
		_ = state.Remove(worktreeName)
		if envResult != nil {
			envResult.Cleanup()
		}
	}

	// Wait for server ready
	fmt.Printf("⏳  Waiting for localhost:%d...\n", p)

	readyCh := make(chan error, 1)
	go func() { readyCh <- server.WaitReady(p, 90*time.Second) }()

	serverExitCh := make(chan error, 1)
	go func() { serverExitCh <- srv.Wait() }()

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

	// Launch browser and handle lifecycle
	for {
		browserCmd, err := browser.Launch(chromeBin, profileDir, url)
		if err != nil {
			cleanup()
			return err
		}
		entry.BrowserPID = browserCmd.Process.Pid
		_ = state.Add(worktreeName, entry)

		// Fan-in: browser exit or OS signal
		browserDone := make(chan error, 1)
		go func() { browserDone <- browserCmd.Wait() }()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-sigCh:
			fmt.Println("\n🔴  Interrupted — stopping dev server...")
			_ = browserCmd.Process.Kill()
			cleanup()
			return nil

		case <-serverExitCh:
			fmt.Println("\n🔴  Dev server exited — closing browser...")
			_ = browserCmd.Process.Kill()
			cleanup()
			return nil

		case <-browserDone:
			// Chrome closed — ask what to do
			fmt.Println()
			action := promptAfterChromeClosed()
			switch action {
			case "r":
				fmt.Printf("🔄  Relaunching Chrome at %s...\n", url)
				continue // relaunch browser, same server
			case "k":
				fmt.Println("🔴  Stopping dev server...")
				cleanup()
				return nil
			default: // "q" or anything else
				fmt.Println("💤  Dev server kept running on port", p)
				fmt.Printf("    Re-attach with: devbrowser -p %d\n", p)
				_ = state.Remove(worktreeName)
				// Don't cleanup env files — server keeps running
				return nil
			}
		}
	}
}

func promptAfterChromeClosed() string {
	fmt.Println("Chrome closed. What would you like to do?")
	fmt.Println("  [r] Relaunch Chrome  (keep session, cookies, localStorage)")
	fmt.Println("  [k] Kill dev server and exit")
	fmt.Println("  [q] Quit devbrowser  (keep dev server running in background)")
	fmt.Print("> ")

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return strings.TrimSpace(strings.ToLower(scanner.Text()))
	}
	return "k"
}

func launchBrowserCmd(binary, profileDir, url string) (*exec.Cmd, error) {
	return browser.Launch(binary, profileDir, url)
}
