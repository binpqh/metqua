package session

import (
	"context"
	"fmt"
	"os"
	"time"
)

const lockPollInterval = 50 * time.Millisecond

// FileLock is a cross-platform advisory file lock.
// On Linux/macOS it uses flock(2) via golang.org/x/sys/unix.
// On Windows it uses LockFileEx via golang.org/x/sys/windows.
// The platform-specific implementations are in lock_unix.go and lock_windows.go.
type FileLock struct {
	path string
	file *os.File
}

// NewFileLock creates a FileLock for the given path.
// The lock file is created if it does not exist.
func NewFileLock(path string) *FileLock {
	return &FileLock{path: path}
}

// Lock acquires an exclusive lock on the lock file.
// It blocks polling until the lock is acquired or ctx is done.
// Returns ErrLockTimeout when ctx deadline is exceeded before lock acquisition.
func (l *FileLock) Lock(ctx context.Context) error {
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return fmt.Errorf("file lock open %s: %w", l.path, err)
	}
	l.file = f

	for {
		if ctx.Err() != nil {
			_ = f.Close()
			return ErrLockTimeout
		}
		ok, err := tryLock(f)
		if err != nil {
			_ = f.Close()
			return fmt.Errorf("file lock acquire %s: %w", l.path, err)
		}
		if ok {
			return nil
		}
		select {
		case <-ctx.Done():
			_ = f.Close()
			return ErrLockTimeout
		case <-time.After(lockPollInterval):
		}
	}
}

// Unlock releases the lock and closes the file.
func (l *FileLock) Unlock() error {
	if l.file == nil {
		return nil
	}
	if err := unlock(l.file); err != nil {
		_ = l.file.Close()
		l.file = nil
		return fmt.Errorf("file lock release %s: %w", l.path, err)
	}
	err := l.file.Close()
	l.file = nil
	return err
}
