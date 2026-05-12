//go:build !linux

package browser

func KillBrowserWSL(windowsProfileDir string)  {}
func WaitForCloseWSL(windowsProfileDir string) {}
