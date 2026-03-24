//go:build !windows

package session

import (
	"os"

	"golang.org/x/sys/unix"
)

// tryLock attempts a non-blocking exclusive flock on f.
// Returns (true, nil) when the lock was acquired, (false, nil) when
// the file is already locked by another process, or (false, err) on error.
func tryLock(f *os.File) (bool, error) {
	err := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB)
	if err == nil {
		return true, nil
	}
	if err == unix.EWOULDBLOCK || err == unix.EAGAIN {
		return false, nil
	}
	return false, err
}

// unlock releases an flock acquired by tryLock.
func unlock(f *os.File) error {
	return unix.Flock(int(f.Fd()), unix.LOCK_UN)
}
