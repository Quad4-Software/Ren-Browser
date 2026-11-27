// Copyright 2026 The Libc Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The buffers of the Windows port's streams.
//
// As in the UCRT, a stream reads and writes through a buffer of 4096 bytes,
// except stdout and stderr, which are not buffered. setvbuf changes that, for
// stdout too; then printf, puts and the other functions that write to stdout
// all go through its buffer. Unlike in the UCRT, where _IOLBF is _IOFBF, a
// line buffered stream writes its buffer out at every newline. exit writes
// out what the streams hold, _exit and abort do not. fflush of a stream that
// is reading does nothing, as in the UCRT.
//
// A stream's buffer serves either reading or writing at a time. Writing after
// reading drops what was read ahead and moves the file position back to the
// reader's; reading after writing writes the buffer out first. A pipe cannot
// move back, so what was read ahead from it is lost.

package libc // import "modernc.org/libc"

import (
	"bytes"
	"math"
	"os"
	"slices"
	"strings"

	"golang.org/x/sys/windows"
	"modernc.org/libc/errno"
	"modernc.org/libc/stdio"
	"modernc.org/libc/sys/types"
	"modernc.org/libc/unistd"
)

// streamBufSize is the size of the buffer of a stream setvbuf did not size, as
// the UCRT's _INTERNAL_BUFSIZ.
const streamBufSize = 4096

// defaultBufMode returns the buffering of a new stream of fd.
func defaultBufMode(fd int32) int32 {
	if fd == unistd.STDOUT_FILENO || fd == unistd.STDERR_FILENO {
		return stdio.X_IONBF
	}

	return stdio.X_IOFBF
}

// streamAccess returns whether a stream of an fopen mode cannot read and
// whether it cannot write.
func streamAccess(mode string) (noRead, noWrite bool) {
	switch {
	case strings.Contains(mode, "+"):
		return false, false
	case len(mode) != 0 && mode[0] == 'r':
		return false, true
	default:
		return true, false
	}
}

// resetBuf empties both buffers and gives f the default buffering, for a
// stream on a newly opened file.
func (f *file) resetBuf() {
	f.bufMode, f.bufSize = defaultBufMode(f._fd), streamBufSize
	f.rbuf, f.rpos, f.wbuf, f.unget = nil, 0, nil, -1
}

// unread returns how many bytes f has read ahead of the position of its
// reader, a character pushed back by ungetc included.
func (f *file) unread() int {
	n := len(f.rbuf) - f.rpos
	if f.unget >= 0 {
		n++
	}
	return n
}

// rawWrite writes all of b to the file, or fails. Standard output on its
// original handle goes through os.Stdout, which writes UTF-8 to a console as
// such.
func (f *file) rawWrite(b []byte) (int, error) {
	if f._fd == unistd.STDOUT_FILENO && f.Handle == windows.Handle(os.Stdout.Fd()) {
		return os.Stdout.Write(b)
	}

	return f.write(b)
}

// flush writes out what f buffered for writing. The buffer is emptied also
// when that fails, as in the UCRT, and the failure sets the error flag.
func (f *file) flush() error {
	if len(f.wbuf) == 0 {
		return nil
	}

	_, err := f.rawWrite(f.wbuf)
	f.wbuf = f.wbuf[:0]
	if err != nil {
		f.hadErr = true
	}
	return err
}

// toWrite drops what f read ahead and moves the file position back to its
// reader's.
func (f *file) toWrite() {
	if n := f.unread(); n != 0 {
		windows.Seek(f.Handle, -int64(n), 1) // A pipe cannot go back.
	}
	f.rbuf, f.rpos, f.unget = f.rbuf[:0], 0, -1
}

// put writes b to f, through its buffer unless f is not buffered.
func (f *file) put(b []byte) (int, error) {
	if f.noWrite {
		return 0, windows.ERROR_INVALID_HANDLE // EBADF, as in the UCRT
	}

	f.toWrite()
	if f.bufMode == stdio.X_IONBF {
		return f.rawWrite(b)
	}

	if len(f.wbuf)+len(b) > f.bufSize {
		if err := f.flush(); err != nil {
			return 0, err
		}

		if len(b) >= f.bufSize {
			return f.rawWrite(b)
		}
	}

	if f.wbuf == nil {
		f.wbuf = make([]byte, 0, f.bufSize)
	}
	f.wbuf = append(f.wbuf, b...)
	if f.bufMode == stdio.X_IOLBF && bytes.IndexByte(b, '\n') >= 0 {
		if err := f.flush(); err != nil {
			return 0, err
		}
	}

	return len(b), nil
}

// fill reads into the buffer of f what one read returns, after f has written
// out its output. It returns the number of bytes read, 0 at the end of the
// file.
func (f *file) fill() (int, error) {
	if err := f.flush(); err != nil {
		return 0, err
	}

	if cap(f.rbuf) < f.bufSize {
		f.rbuf = make([]byte, f.bufSize)
	}
	n, err := windows.Read(f.Handle, f.rbuf[:f.bufSize])
	f.rbuf, f.rpos = f.rbuf[:n], 0
	return n, err
}

// getc returns the next byte of f, or -1 at the end of the file, where it sets
// the end of file flag, and on an error, which it returns.
func (f *file) getc() (int32, error) {
	if c := f.unget; c >= 0 {
		f.unget = -1
		return c, nil
	}

	if f.rpos < len(f.rbuf) {
		f.rpos++
		return int32(f.rbuf[f.rpos-1]), nil
	}

	var b [1]byte
	var n int
	var err error
	switch {
	case f.bufMode == stdio.X_IONBF:
		if err = f.flush(); err == nil {
			n, err = windows.Read(f.Handle, b[:])
		}
	default:
		if n, err = f.fill(); n != 0 {
			b[0] = f.rbuf[0]
			f.rpos = 1
		}
	}
	if err != nil {
		return -1, err
	}

	if n == 0 {
		f.eof = true
		return -1, nil
	}

	return int32(b[0]), nil
}

// read reads into b until it is full, the end of the file or an error. A pipe
// or the console can return less than asked for before the end.
func (f *file) read(b []byte) (n int, err error) {
	if c := f.unget; c >= 0 && len(b) != 0 {
		f.unget = -1
		b[0] = byte(c)
		n = 1
	}
	m := copy(b[n:], f.rbuf[f.rpos:])
	f.rpos += m
	n += m
	if n < len(b) {
		if err := f.flush(); err != nil {
			return n, err
		}
	}

	for n < len(b) {
		if f.bufMode == stdio.X_IONBF || len(b)-n >= f.bufSize {
			if m, err = windows.Read(f.Handle, b[n:]); err == nil && m != 0 {
				n += m
				continue
			}
		} else if m, err = f.fill(); err == nil && m != 0 {
			f.rpos = copy(b[n:], f.rbuf)
			n += f.rpos
			continue
		}

		if err == nil {
			f.eof = true
		}
		return n, err
	}
	return n, nil
}

// flushAll writes out the output buffered by all streams, in the order of
// their descriptors, and returns the first failure. At exit it skips the
// streams that another thread is using, which may be blocked reading.
func flushAll(atExit bool) (err error) {
	w_fdLock.Lock()
	fds := make([]int32, 0, len(w_fd_to_file))
	for fd := range w_fd_to_file {
		fds = append(fds, fd)
	}
	slices.Sort(fds)
	fs := make([]*file, len(fds))
	for i, fd := range fds {
		fs[i] = w_fd_to_file[fd]
	}
	w_fdLock.Unlock()
	for _, f := range fs {
		switch {
		case atExit:
			if !f.mu.TryLock() {
				continue
			}
		default:
			f.mu.Lock()
		}
		if e := f.flush(); e != nil && err == nil {
			err = e
		}
		f.mu.Unlock()
	}
	return err
}

// flushStreams writes out the output buffered by all streams, for exit.
func flushStreams() {
	flushAll(true)
}

// flushFile writes out the output buffered by f, before f leaves the
// descriptor table.
func flushFile(f *file) {
	f.mu.Lock()
	f.flush()
	f.mu.Unlock()
}

// write writes b to standard output, for printf and puts.
func write(b []byte) (int, error) {
	n, err := fwrite(unistd.STDOUT_FILENO, b)
	if err != nil {
		return -1, err
	}

	return n, nil
}

// fwrite writes b to the stream of fd, which is stdout or stderr when fd is
// not open.
func fwrite(fd int32, b []byte) (int, error) {
	f, ok := fdToFile(fd)
	if !ok {
		switch fd {
		case unistd.STDOUT_FILENO:
			return os.Stdout.Write(b)
		case unistd.STDERR_FILENO:
			return os.Stderr.Write(b)
		}

		return -1, windows.ERROR_INVALID_HANDLE
	}

	if dmesgs {
		dmesg("%v: fd %v: %s", origin(1), fd, b)
	}
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.put(b)
}

// int setvbuf(FILE *stream, char *buf, int mode, size_t size);
//
// setvbuf writes out what the stream buffered and sets its buffering. It
// takes effect at any time, as in the UCRT, and keeps what the stream read
// ahead. The stream uses a buffer of its own of the size, not buf.
func Xsetvbuf(t *TLS, stream, buf uintptr, mode int32, size types.Size_t) int32 {
	if __ccgo_strace {
		trc("t=%v buf=%v mode=%v size=%v, (%v:)", t, buf, mode, size, origin(2))
	}
	f, ok := winGetObject(stream).(*file)
	if !ok {
		t.setErrno(errno.EINVAL)
		return -1
	}

	switch mode {
	case stdio.X_IONBF:
	case stdio.X_IOFBF, stdio.X_IOLBF:
		if size < 2 || uint64(size) > math.MaxInt32 {
			t.setErrno(errno.EINVAL)
			return -1
		}
	default:
		t.setErrno(errno.EINVAL)
		return -1
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.flush()
	f.wbuf, f.bufMode = nil, mode
	if mode != stdio.X_IONBF {
		f.bufSize = int(size)
	}
	return 0
}

// setbuf is setbuf(3): no buffering for a NULL buf, else full buffering with
// BUFSIZ bytes.
func setbuf(t *TLS, stream, buf uintptr) {
	if buf == 0 {
		Xsetvbuf(t, stream, 0, stdio.X_IONBF, 0)
		return
	}

	Xsetvbuf(t, stream, buf, stdio.X_IOFBF, stdio.BUFSIZ)
}
