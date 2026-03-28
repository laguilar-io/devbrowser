//go:build linux

package browser

var candidates = []string{
	"/usr/bin/google-chrome-stable",
	"/usr/bin/google-chrome",
	"/usr/bin/chromium-browser",
	"/usr/bin/chromium",
	"/snap/bin/chromium",
}
