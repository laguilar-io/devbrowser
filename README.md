# devbrowser

**One command. Isolated Chrome per worktree. DevTools always open.**

```
devbrowser feature-auth
```

↓

```
  devbrowser

  worktree  ~/dev/myapp/.worktrees/feature-auth
  command   pnpm run dev
  port      3001
  url       http://localhost:3001
  profile   ~/.devbrowser/profiles/myapp__feature-auth
```

Chrome opens on port 3001 with its own cookies, localStorage, and login session — completely isolated from every other tab or worktree. DevTools is open automatically. When you close Chrome, the dev server stops.

---

## The problem

You're running three worktrees in parallel. You open Chrome for each one. By the third tab, you've been logged out of the first two, localStorage from one feature is leaking into another, and you've accidentally tested the wrong build twice.

**devbrowser fixes this.** Each worktree gets its own Chrome profile — a fully separate browser identity. No session bleeding. No shared cookies. No shared localStorage. Ever.

---

## Install

```bash
go install github.com/laguilar-io/devbrowser/cmd/devbrowser@latest
```

Or download a binary from [Releases](https://github.com/laguilar-io/devbrowser/releases) (macOS, Linux, Windows — amd64/arm64).

---

## Usage

```bash
# Start dev server + Chrome for a worktree
devbrowser feature-auth

# Use a custom dev command
devbrowser feature-auth -c "npm run dev"

# Override the port
devbrowser feature-auth -p 3005

# From inside a worktree directory (no arg needed)
devbrowser

# Attach Chrome to an already-running server
devbrowser -p 3001

# List active sessions
devbrowser list

# Stop a session
devbrowser stop feature-auth
devbrowser stop --all
```

When Chrome closes, devbrowser asks what to do:

```
Chrome closed. What would you like to do?
  [r] Relaunch Chrome  (keeps session, cookies, localStorage)
  [k] Kill dev server and exit
  [q] Quit devbrowser  (keep dev server running in background)
```

Pick **r** to get Chrome back without losing any state. Pick **q** to keep the dev server alive and come back later — running `devbrowser feature-auth` again will reattach without starting a new server.

---

## How it works

1. Resolves the worktree path via `git worktree list`
2. Finds the next available port (starting from 3000)
3. Injects `PORT=<n>` into the environment — works with Next.js, Vite, CRA, and anything that respects `$PORT`
4. Runs your dev command inside the worktree directory
5. Polls the port every 200ms until the server is ready (up to 90s)
6. Launches Chrome with:
   - `--user-data-dir=~/.devbrowser/profiles/<worktree>` — isolated profile, persisted between sessions
   - `--disable-extensions` — clean, predictable environment
   - `--auto-open-devtools-for-tabs` — DevTools open on every new tab
7. On macOS, detects real window close (not just CMD+H) via osascript
8. Copies `.env*.local` files from the repo root to the worktree automatically

Session state (PID, port, profile path) is saved in `~/.devbrowser/state.json` so reattach always uses the right profile for the right worktree.

---

## Config

`~/.devbrowser/config.toml` is created on first run with sensible defaults:

```toml
default_command = "pnpm run dev"
start_port      = 3000
profiles_dir    = ""          # default: ~/.devbrowser/profiles
browser_path    = ""          # default: auto-detect Chrome/Chromium
```

---

## Requirements

- Git
- Chrome or Chromium (auto-detected on macOS, Linux, Windows)
- Go 1.21+ (only if installing via `go install`)

---

## Why not just open a new Chrome window?

A regular Chrome window shares everything: cookies, localStorage, IndexedDB, service workers. If you log into your app in one tab, you're logged into every tab. If your staging environment shares a domain with your local dev, they fight over the same session.

`--user-data-dir` creates a completely separate browser identity. It's the same mechanism Playwright and Puppeteer use for test isolation — devbrowser brings that to your daily development workflow.

---

## License

MIT
