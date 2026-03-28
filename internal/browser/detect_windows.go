//go:build windows

package browser

import "os"

var candidates = []string{
	os.Getenv("LOCALAPPDATA") + `\Google\Chrome\Application\chrome.exe`,
	os.Getenv("PROGRAMFILES") + `\Google\Chrome\Application\chrome.exe`,
	os.Getenv("PROGRAMFILES(X86)") + `\Google\Chrome\Application\chrome.exe`,
}
