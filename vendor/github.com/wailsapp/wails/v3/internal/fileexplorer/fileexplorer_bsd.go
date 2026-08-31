//go:build freebsd || netbsd || openbsd

package fileexplorer

import (
	"fmt"
	"runtime"
	"syscall"
)

func explorerBinArgs(path string, selectFile bool) (string, []string, error) {
	return "", nil, fmt.Errorf("file explorer not supported on %s", runtime.GOOS)
}

func sysProcAttr(path string, selectFile bool) *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}
