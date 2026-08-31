//go:build freebsd || netbsd || openbsd

package operatingsystem

import "runtime"

func platformInfo() (*OS, error) {
	return &OS{
		ID:      runtime.GOOS,
		Name:    runtime.GOOS,
		Version: "unknown",
	}, nil
}
