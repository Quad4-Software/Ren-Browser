// SPDX-License-Identifier: MIT

//go:build linux

package sandbox

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"unsafe"

	"github.com/adrg/xdg"
	"golang.org/x/sys/unix"
)

// landlockAccessFSResolveUnix is ABI v9 (not yet in all golang.org/x/sys snapshots).
const landlockAccessFSResolveUnix = uint64(1) << 16

var (
	kernelSupportOnce sync.Once
	kernelSupported   bool
	abiOnce           sync.Once
	cachedABI         int
)

// KernelSupported reports whether this kernel exposes Landlock (5.13+).
func KernelSupported() bool {
	kernelSupportOnce.Do(func() {
		kernelSupported = ABIVersion() > 0
	})
	return kernelSupported
}

// ABIVersion returns the Landlock ABI version, or 0 when unavailable.
func ABIVersion() int {
	abiOnce.Do(func() {
		cachedABI = probeABI()
	})
	return cachedABI
}

func probeABI() int {
	abi, _, errno := unix.Syscall(unix.SYS_LANDLOCK_CREATE_RULESET,
		0, 0, unix.LANDLOCK_CREATE_RULESET_VERSION)
	if errno == unix.ENOSYS || errno == unix.EOPNOTSUPP {
		return 0
	}
	if errno != 0 || int(abi) <= 0 {
		return 0
	}
	return int(abi)
}

func applyLandlock(opts Options) error {
	abi := ABIVersion()
	if abi <= 0 {
		return fmt.Errorf("landlock not supported by kernel")
	}

	handledFS, scoped := landlockHandledForABI(abi)
	attr := unix.LandlockRulesetAttr{
		Access_fs: handledFS,
		Scoped:    scoped,
	}

	fd, _, errno := unix.Syscall(unix.SYS_LANDLOCK_CREATE_RULESET,
		uintptr(unsafe.Pointer(&attr)), // #nosec G103 -- required for direct Landlock syscall interface
		uintptr(unsafe.Sizeof(attr)),
		0)
	if errno == unix.ENOSYS || errno == unix.EOPNOTSUPP {
		return fmt.Errorf("landlock not supported by kernel")
	}
	if errno != 0 {
		return fmt.Errorf("landlock_create_ruleset: %w", errno)
	}
	rulesetFD := int(fd) // #nosec G115 -- syscall fd is always a small non-negative integer on Linux
	defer unix.Close(rulesetFD)

	for _, root := range collectReadRoots(opts) {
		_ = landlockAddRule(rulesetFD, root, landlockReadAccess(handledFS), handledFS)
	}

	for _, root := range collectWriteRoots(opts) {
		_ = landlockAddRule(rulesetFD, root, handledFS, handledFS)
	}

	for _, rule := range landlockSystemFiles(handledFS) {
		_ = landlockAddRule(rulesetFD, rule.path, rule.access, handledFS)
	}

	flags := uintptr(0)
	if abi >= 8 {
		flags |= unix.LANDLOCK_RESTRICT_SELF_TSYNC
	}

	_, _, errno = unix.Syscall(unix.SYS_LANDLOCK_RESTRICT_SELF,
		uintptr(rulesetFD), // #nosec G115 -- converting syscall fd to uintptr for raw syscall
		flags, 0)
	if errno != 0 {
		return fmt.Errorf("landlock_restrict_self: %w", errno)
	}
	return nil
}

func landlockHandledForABI(abi int) (accessFS, scoped uint64) {
	accessFS = uint64(
		unix.LANDLOCK_ACCESS_FS_EXECUTE |
			unix.LANDLOCK_ACCESS_FS_WRITE_FILE |
			unix.LANDLOCK_ACCESS_FS_READ_FILE |
			unix.LANDLOCK_ACCESS_FS_READ_DIR |
			unix.LANDLOCK_ACCESS_FS_REMOVE_DIR |
			unix.LANDLOCK_ACCESS_FS_REMOVE_FILE |
			unix.LANDLOCK_ACCESS_FS_MAKE_CHAR |
			unix.LANDLOCK_ACCESS_FS_MAKE_DIR |
			unix.LANDLOCK_ACCESS_FS_MAKE_REG |
			unix.LANDLOCK_ACCESS_FS_MAKE_SOCK |
			unix.LANDLOCK_ACCESS_FS_MAKE_FIFO |
			unix.LANDLOCK_ACCESS_FS_MAKE_BLOCK |
			unix.LANDLOCK_ACCESS_FS_MAKE_SYM |
			unix.LANDLOCK_ACCESS_FS_REFER |
			unix.LANDLOCK_ACCESS_FS_TRUNCATE |
			unix.LANDLOCK_ACCESS_FS_IOCTL_DEV |
			landlockAccessFSResolveUnix,
	)
	scoped = uint64(
		unix.LANDLOCK_SCOPE_ABSTRACT_UNIX_SOCKET |
			unix.LANDLOCK_SCOPE_SIGNAL,
	)

	// Compatibility fallthrough matches kernel userspace docs.
	switch {
	case abi < 2:
		accessFS &^= unix.LANDLOCK_ACCESS_FS_REFER
		fallthrough
	case abi < 3:
		accessFS &^= unix.LANDLOCK_ACCESS_FS_TRUNCATE
		fallthrough
	case abi < 5:
		accessFS &^= unix.LANDLOCK_ACCESS_FS_IOCTL_DEV
		fallthrough
	case abi < 6:
		scoped = 0
		fallthrough
	case abi < 9:
		accessFS &^= landlockAccessFSResolveUnix
	}

	return accessFS, scoped
}

func landlockReadAccess(handledFS uint64) uint64 {
	return handledFS & uint64(
		unix.LANDLOCK_ACCESS_FS_EXECUTE|
			unix.LANDLOCK_ACCESS_FS_READ_FILE|
			unix.LANDLOCK_ACCESS_FS_READ_DIR|
			landlockAccessFSResolveUnix,
	)
}

func landlockReadOnlyFile(handledFS uint64) uint64 {
	return handledFS & uint64(
		unix.LANDLOCK_ACCESS_FS_EXECUTE|
			unix.LANDLOCK_ACCESS_FS_READ_FILE|
			unix.LANDLOCK_ACCESS_FS_IOCTL_DEV|
			landlockAccessFSResolveUnix,
	)
}

type landlockRule struct {
	path   string
	access uint64
}

func landlockSystemFiles(handledFS uint64) []landlockRule {
	roFile := landlockReadOnlyFile(handledFS)
	roDir := landlockReadAccess(handledFS)
	return []landlockRule{
		{"/etc/resolv.conf", roFile},
		{"/etc/hosts", roFile},
		{"/etc/ssl/cert.pem", roFile},
		{"/etc/ssl/certs", roDir},
		{"/etc/ca-certificates", roDir},
		{"/proc/self", roFile},
		{"/dev/null", roFile},
		{"/dev/urandom", roFile},
		{"/dev/dri", roDir},
	}
}

func collectReadRoots(opts Options) []string {
	roots := map[string]struct{}{
		"/usr":   {},
		"/lib":   {},
		"/lib64": {},
		"/etc":   {},
		"/bin":   {},
		"/sbin":  {},
		"/proc":  {},
		"/sys":   {},
		"/opt":   {},
	}
	for _, path := range opts.ExtraReadPaths {
		if dir := existingDir(path); dir != "" {
			roots[dir] = struct{}{}
		}
	}
	out := make([]string, 0, len(roots))
	for root := range roots {
		if _, err := os.Stat(root); err == nil {
			out = append(out, root)
		}
	}
	return out
}

func collectWriteRoots(opts Options) []string {
	candidates := []string{
		opts.DataDir,
		opts.ReticulumDir,
		filepath.Dir(opts.ReticulumConfig),
		opts.PluginsDir,
		opts.DownloadDir,
		opts.AssetsDir,
		filepath.Dir(opts.AssetsZip),
		"/tmp",
		"/var/tmp",
		"/dev/shm",
		"/run",
		stringsTrim(xdg.CacheHome),
		stringsTrim(xdg.ConfigHome),
		stringsTrim(xdg.DataHome),
		stringsTrim(xdg.StateHome),
		stringsTrim(os.Getenv("XDG_RUNTIME_DIR")),
	}
	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		candidates = append(candidates,
			filepath.Join(home, ".cache"),
			filepath.Join(home, ".config"),
			filepath.Join(home, ".local", "share"),
			filepath.Join(home, ".local", "state"),
		)
	}

	seen := make(map[string]struct{})
	var out []string
	for _, candidate := range candidates {
		dir := existingDir(candidate)
		if dir == "" {
			continue
		}
		if _, ok := seen[dir]; ok {
			continue
		}
		seen[dir] = struct{}{}
		out = append(out, dir)
	}
	if _, err := os.Stat("/dev"); err == nil {
		out = append(out, "/dev")
	}
	return out
}

func existingDir(path string) string {
	path = stringsTrim(path)
	if path == "" {
		return ""
	}
	resolved := filepath.Clean(path)
	if info, err := os.Stat(resolved); err == nil {
		if info.IsDir() {
			return resolved
		}
		parent := filepath.Dir(resolved)
		if parent != "" && parent != resolved {
			if parentInfo, parentErr := os.Stat(parent); parentErr == nil && parentInfo.IsDir() {
				return parent
			}
		}
		return ""
	}
	parent := filepath.Dir(resolved)
	if parent != "" && parent != resolved {
		if parentInfo, err := os.Stat(parent); err == nil && parentInfo.IsDir() {
			return parent
		}
	}
	return ""
}

func landlockAddRule(rulesetFD int, path string, access, handledFS uint64) error {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		if os.IsNotExist(err) {
			return unix.ENOENT
		}
		return err
	}

	fd, err := unix.Open(resolved, unix.O_PATH|unix.O_CLOEXEC, 0)
	if err != nil {
		if os.IsNotExist(err) {
			return unix.ENOENT
		}
		return err
	}
	defer unix.Close(fd)

	info, err := os.Stat(resolved)
	if err != nil {
		return err
	}
	allowed := access & handledFS
	if !info.IsDir() {
		allowed &= uint64(
			unix.LANDLOCK_ACCESS_FS_READ_FILE |
				unix.LANDLOCK_ACCESS_FS_WRITE_FILE |
				unix.LANDLOCK_ACCESS_FS_TRUNCATE |
				unix.LANDLOCK_ACCESS_FS_EXECUTE |
				unix.LANDLOCK_ACCESS_FS_IOCTL_DEV |
				landlockAccessFSResolveUnix,
		)
		allowed &= handledFS
	}
	if allowed == 0 {
		return nil
	}

	attr := unix.LandlockPathBeneathAttr{
		Allowed_access: allowed,
		Parent_fd:      int32(fd), // #nosec G115 -- O_PATH fd from unix.Open is always small non-negative
	}

	_, _, errno := unix.Syscall(unix.SYS_LANDLOCK_ADD_RULE,
		uintptr(rulesetFD), // #nosec G115 -- converting syscall fd to uintptr for raw syscall
		uintptr(unix.LANDLOCK_RULE_PATH_BENEATH),
		uintptr(unsafe.Pointer(&attr))) // #nosec G103 -- required for direct Landlock syscall interface
	if errno != 0 {
		return errno
	}
	return nil
}
