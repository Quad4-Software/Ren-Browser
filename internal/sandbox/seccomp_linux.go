// SPDX-License-Identifier: MIT

//go:build linux

package sandbox

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"golang.org/x/sys/unix"
)

var (
	seccompSupportOnce sync.Once
	seccompSupported   bool
)

// SeccompSupported reports whether classic seccomp-bpf filters can be installed.
func SeccompSupported() bool {
	seccompSupportOnce.Do(func() {
		seccompSupported = probeSeccomp()
	})
	return seccompSupported
}

func probeSeccomp() bool {
	// Query whether SECCOMP_SET_MODE_FILTER is available without installing a filter.
	_, _, errno := unix.Syscall(unix.SYS_SECCOMP,
		unix.SECCOMP_GET_ACTION_AVAIL,
		0,
		uintptr(unsafe.Pointer(&[]uint32{unix.SECCOMP_RET_ALLOW}[0]))) // #nosec G103 -- kernel ABI probe
	if errno == unix.ENOSYS || errno == unix.EOPNOTSUPP {
		return false
	}
	// EINVAL/EFAULT still mean the seccomp syscall exists.
	return true
}

func applySeccomp() error {
	if !SeccompSupported() {
		return fmt.Errorf("seccomp not supported by kernel")
	}

	filter, err := buildSeccompFilter()
	if err != nil {
		return err
	}
	prog := unix.SockFprog{
		Len:    uint16(len(filter)), // #nosec G115 -- filter length is tiny and fits uint16
		Filter: &filter[0],
	}

	flags := uintptr(unix.SECCOMP_FILTER_FLAG_TSYNC)
	_, _, errno := unix.Syscall(unix.SYS_SECCOMP,
		unix.SECCOMP_SET_MODE_FILTER,
		flags,
		uintptr(unsafe.Pointer(&prog))) // #nosec G103 -- required for seccomp filter install
	if errno == unix.ENOSYS || errno == unix.EOPNOTSUPP {
		return fmt.Errorf("seccomp not supported by kernel")
	}
	if errno == unix.EINVAL {
		// Older kernels may lack TSYNC. Retry for the calling thread only.
		_, _, errno = unix.Syscall(unix.SYS_SECCOMP,
			unix.SECCOMP_SET_MODE_FILTER,
			0,
			uintptr(unsafe.Pointer(&prog))) // #nosec G103 -- required for seccomp filter install
	}
	if errno != 0 {
		return fmt.Errorf("seccomp filter: %w", errno)
	}
	return nil
}

func buildSeccompFilter() ([]unix.SockFilter, error) {
	arch, err := seccompAuditArch()
	if err != nil {
		return nil, err
	}
	blocked := blockedSyscalls()
	if len(blocked) == 0 {
		return nil, fmt.Errorf("no syscalls to block on this architecture")
	}

	const (
		seccompDataNr   = 0
		seccompDataArch = 4
	)

	deny := uint32(unix.SECCOMP_RET_ERRNO) | uint32(unix.EPERM)
	allow := uint32(unix.SECCOMP_RET_ALLOW)

	filter := []unix.SockFilter{
		// Validate architecture.
		{Code: unix.BPF_LD | unix.BPF_W | unix.BPF_ABS, K: seccompDataArch},
		{Code: unix.BPF_JMP | unix.BPF_JEQ | unix.BPF_K, Jt: 1, Jf: 0, K: arch},
		{Code: unix.BPF_RET | unix.BPF_K, K: uint32(unix.SECCOMP_RET_KILL_PROCESS)},
		// Load syscall number.
		{Code: unix.BPF_LD | unix.BPF_W | unix.BPF_ABS, K: seccompDataNr},
	}

	for _, nr := range blocked {
		// jeq nr -> deny, else continue
		filter = append(filter,
			unix.SockFilter{Code: unix.BPF_JMP | unix.BPF_JEQ | unix.BPF_K, Jt: 0, Jf: 1, K: uint32(nr)}, // #nosec G115 -- syscall numbers fit uint32
			unix.SockFilter{Code: unix.BPF_RET | unix.BPF_K, K: deny},
		)
	}
	filter = append(filter, unix.SockFilter{Code: unix.BPF_RET | unix.BPF_K, K: allow})
	return filter, nil
}

func seccompAuditArch() (uint32, error) {
	switch runtime.GOARCH {
	case "amd64":
		return unix.AUDIT_ARCH_X86_64, nil
	case "arm64":
		return unix.AUDIT_ARCH_AARCH64, nil
	case "arm":
		return unix.AUDIT_ARCH_ARM, nil
	case "386":
		return unix.AUDIT_ARCH_I386, nil
	case "ppc64le":
		return unix.AUDIT_ARCH_PPC64LE, nil
	case "ppc64":
		return unix.AUDIT_ARCH_PPC64, nil
	case "ppc":
		return unix.AUDIT_ARCH_PPC, nil
	case "riscv64":
		return unix.AUDIT_ARCH_RISCV64, nil
	case "s390x":
		return unix.AUDIT_ARCH_S390X, nil
	case "mips64":
		return unix.AUDIT_ARCH_MIPS64, nil
	case "mips64le":
		return unix.AUDIT_ARCH_MIPSEL64, nil
	case "loong64":
		return unix.AUDIT_ARCH_LOONGARCH64, nil
	default:
		return 0, fmt.Errorf("seccomp unsupported GOARCH %q", runtime.GOARCH)
	}
}

func blockedSyscalls() []uint32 {
	// Denylist of high-risk privileged operations. Prefer errno over kill so
	// unexpected library probes fail soft instead of aborting the process.
	candidates := []uintptr{
		unix.SYS_MOUNT,
		unix.SYS_UMOUNT2,
		unix.SYS_PIVOT_ROOT,
		unix.SYS_SWAPON,
		unix.SYS_SWAPOFF,
		unix.SYS_REBOOT,
		unix.SYS_ACCT,
		unix.SYS_INIT_MODULE,
		unix.SYS_FINIT_MODULE,
		unix.SYS_DELETE_MODULE,
		unix.SYS_ADD_KEY,
		unix.SYS_REQUEST_KEY,
		unix.SYS_KEYCTL,
		unix.SYS_BPF,
		unix.SYS_PERF_EVENT_OPEN,
		unix.SYS_USERFAULTFD,
		unix.SYS_UNSHARE,
		unix.SYS_SETNS,
		unix.SYS_KEXEC_LOAD,
		unix.SYS_KEXEC_FILE_LOAD,
		unix.SYS_OPEN_BY_HANDLE_AT,
		unix.SYS_FSOPEN,
		unix.SYS_FSCONFIG,
		unix.SYS_FSMOUNT,
		unix.SYS_FSPICK,
		unix.SYS_OPEN_TREE,
		unix.SYS_MOVE_MOUNT,
		unix.SYS_MOUNT_SETATTR,
		unix.SYS_IO_URING_SETUP,
		unix.SYS_IO_URING_ENTER,
		unix.SYS_IO_URING_REGISTER,
	}

	out := make([]uint32, 0, len(candidates))
	seen := make(map[uint32]struct{}, len(candidates))
	for _, nr := range candidates {
		if nr == 0 {
			continue
		}
		u := uint32(nr) // #nosec G115 -- Linux syscall numbers fit uint32
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	return out
}
