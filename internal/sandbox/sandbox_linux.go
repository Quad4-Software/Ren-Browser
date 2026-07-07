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

var (
	kernelSupportOnce sync.Once
	kernelSupported   bool
)

// KernelSupported reports whether this kernel exposes Landlock (5.13+).
func KernelSupported() bool {
	kernelSupportOnce.Do(func() {
		kernelSupported = probeLandlock()
	})
	return kernelSupported
}

func probeLandlock() bool {
	attr := unix.LandlockRulesetAttr{
		Access_fs: landlockAccessFS,
	}
	fd, _, errno := unix.Syscall(unix.SYS_LANDLOCK_CREATE_RULESET,
		uintptr(unsafe.Pointer(&attr)),
		uintptr(unsafe.Sizeof(attr)),
		0)
	if errno == unix.ENOSYS || errno == unix.EOPNOTSUPP {
		return false
	}
	if errno != 0 {
		return false
	}
	_ = unix.Close(int(fd))
	return true
}

func applyPlatform(opts Options) error {
	if !opts.ServerMode {
		if os.Getenv("WEBKIT_DISABLE_SANDBOX_THIS_IS_DANGEROUS") == "" {
			_ = os.Setenv("WEBKIT_DISABLE_SANDBOX_THIS_IS_DANGEROUS", "1")
		}
	}
	if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0); err != nil {
		return fmt.Errorf("PR_SET_NO_NEW_PRIVS: %w", err)
	}
	return applyLandlock(opts)
}

func applyLandlock(opts Options) error {
	attr := unix.LandlockRulesetAttr{
		Access_fs: landlockAccessFS,
	}

	fd, _, errno := unix.Syscall(unix.SYS_LANDLOCK_CREATE_RULESET,
		uintptr(unsafe.Pointer(&attr)),
		uintptr(unsafe.Sizeof(attr)),
		0)
	if errno == unix.ENOSYS {
		return fmt.Errorf("landlock not supported by kernel")
	}
	if errno != 0 {
		return fmt.Errorf("landlock_create_ruleset: %w", errno)
	}
	rulesetFD := int(fd)
	defer unix.Close(rulesetFD)

	for _, root := range collectReadRoots(opts) {
		if err := landlockAddRule(rulesetFD, root, landlockReadAccess); err != nil {
			if err != unix.ENOENT {
				continue
			}
		}
	}

	for _, root := range collectWriteRoots(opts) {
		if err := landlockAddRule(rulesetFD, root, landlockAccessFS); err != nil {
			if err != unix.ENOENT {
				continue
			}
		}
	}

	for _, rule := range landlockSystemFiles {
		if err := landlockAddRule(rulesetFD, rule.path, rule.access); err != nil && err != unix.ENOENT {
			continue
		}
	}

	_, _, errno = unix.Syscall(unix.SYS_LANDLOCK_RESTRICT_SELF,
		uintptr(rulesetFD),
		0, 0)
	if errno != 0 {
		return fmt.Errorf("landlock_restrict_self: %w", errno)
	}
	return nil
}

type landlockRule struct {
	path   string
	access uint64
}

var landlockAccessFS = uint64(
	unix.LANDLOCK_ACCESS_FS_READ_FILE |
		unix.LANDLOCK_ACCESS_FS_READ_DIR |
		unix.LANDLOCK_ACCESS_FS_WRITE_FILE |
		unix.LANDLOCK_ACCESS_FS_REMOVE_FILE |
		unix.LANDLOCK_ACCESS_FS_REMOVE_DIR |
		unix.LANDLOCK_ACCESS_FS_MAKE_CHAR |
		unix.LANDLOCK_ACCESS_FS_MAKE_DIR |
		unix.LANDLOCK_ACCESS_FS_MAKE_REG |
		unix.LANDLOCK_ACCESS_FS_MAKE_SOCK |
		unix.LANDLOCK_ACCESS_FS_MAKE_FIFO |
		unix.LANDLOCK_ACCESS_FS_MAKE_BLOCK |
		unix.LANDLOCK_ACCESS_FS_MAKE_SYM |
		unix.LANDLOCK_ACCESS_FS_TRUNCATE |
		unix.LANDLOCK_ACCESS_FS_REFER,
)

var landlockReadAccess = uint64(
	unix.LANDLOCK_ACCESS_FS_READ_FILE |
		unix.LANDLOCK_ACCESS_FS_READ_DIR |
		unix.LANDLOCK_ACCESS_FS_EXECUTE,
)

var landlockReadOnlyFile = uint64(unix.LANDLOCK_ACCESS_FS_READ_FILE | unix.LANDLOCK_ACCESS_FS_EXECUTE)

var landlockSystemFiles = []landlockRule{
	{"/etc/resolv.conf", landlockReadOnlyFile},
	{"/etc/hosts", landlockReadOnlyFile},
	{"/etc/ssl/cert.pem", landlockReadOnlyFile},
	{"/etc/ssl/certs", landlockReadAccess},
	{"/etc/ca-certificates", landlockReadAccess},
	{"/proc/self", landlockReadOnlyFile},
	{"/dev/null", landlockReadOnlyFile},
	{"/dev/urandom", landlockReadOnlyFile},
	{"/dev/dri", landlockReadAccess},
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

func landlockAddRule(rulesetFD int, path string, access uint64) error {
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
	allowed := access
	if !info.IsDir() {
		allowed &= uint64(
			unix.LANDLOCK_ACCESS_FS_READ_FILE |
				unix.LANDLOCK_ACCESS_FS_WRITE_FILE |
				unix.LANDLOCK_ACCESS_FS_TRUNCATE |
				unix.LANDLOCK_ACCESS_FS_EXECUTE,
		)
	}

	attr := unix.LandlockPathBeneathAttr{
		Allowed_access: allowed,
		Parent_fd:      int32(fd),
	}

	_, _, errno := unix.Syscall(unix.SYS_LANDLOCK_ADD_RULE,
		uintptr(rulesetFD),
		uintptr(unix.LANDLOCK_RULE_PATH_BENEATH),
		uintptr(unsafe.Pointer(&attr)))
	if errno != 0 {
		return errno
	}
	return nil
}
