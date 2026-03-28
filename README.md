# devbrowser

> One command to launch an isolated Chrome session per git worktree.

`devbrowser` bridges git worktrees with browser sessions so you can work on multiple features in parallel without cookies, sessions, or localStorage bleeding between them.

```
devbrowser feature-login
```

That's it. devbrowser finds your worktree, picks the next free port, starts your dev server, waits for it to be ready, and opens Chrome with its own isolated profile. When you close Chrome, the dev server stops.

## Why

Modern development with AI agents means running multiple git worktrees simultaneously. Each worktree needs its own dev server on its own port, and its own browser session so you can be logged in as different users, test different states, and not have one tab overwrite another's localStorage.

No existing tool closed this loop. devbrowser does.

## Install

### Homebrew (macOS/Linux)
```bash
brew install laguilar-io/tap/devbrowser
```

### go install
```bash
go install github.com/laguilar-io/devbrowser/cmd/devbrowser@latest
```

### Binary
Download from [Releases](https://github.com/laguilar-io/devbrowser/releases).

## Usage

```bash
# Start a worktree's dev server and open Chrome
devbrowser <worktree-name>

# Same, with a custom command
devbrowser <worktree-name> -c "npm run dev"

# Attach browser only (server already running)
devbrowser
devbrowser -p 3001

# List all active sessions
devbrowser list

# Stop a session
devbrowser stop <worktree-name>
devbrowser stop --all
```

## How it works

1. Finds your worktree via `git worktree list`
2. Scans for the next available port starting from 3000
3. Runs your dev command (`pnpm run dev -p <port>`) inside the worktree
4. Waits for the server to be ready (polling the port)
5. Launches Chrome with:
   - `--user-data-dir=~/.devbrowser/profiles/<worktree-name>` — isolated profile
   - `--disable-extensions` — clean environment
   - `--auto-open-devtools-for-tabs` — DevTools always open
6. When Chrome closes → dev server is stopped automatically

## Config

`~/.devbrowser/config.toml` is created on first run:

```toml
default_command = "pnpm run dev"
start_port = 3000
browser = "auto"
profiles_dir = "~/.devbrowser/profiles"
browser_path = ""  # optional: override Chrome binary path
```

## Requirements

- Git
- Chrome or Chromium
- macOS, Linux, or Windows

## License

MIT
