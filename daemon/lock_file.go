package daemon

import (
	"errors"
	"fmt"
	"os"
)

// ErrWouldBlock resource unavailble error
var ErrWouldBlock = errors.New("daemon: Resource temporarily unavailable")

// LockFile lock file
type LockFile struct {
	*os.File
}

// NewLockFile new lock file
func NewLockFile(file *os.File) *LockFile {
	return &LockFile{file}
}

// OpenLockFile open locked file
func OpenLockFile(name string, perm os.FileMode) (lock *LockFile, err error) {
	var file *os.File
	if file, err = os.OpenFile(name, os.O_RDWR|os.O_CREATE, perm); err == nil {
		lock = &LockFile{file}
	}
	return
}

// Lock lock file
func (file *LockFile) Lock() error {
	return lockFile(file.Fd())
}

// Unlock unlock file
func (file *LockFile) Unlock() error {
	return unlockFile(file.Fd())
}

// WritePid write pid file
func (file *LockFile) WritePid() (err error) {
	if _, err = file.Seek(0, os.SEEK_SET); err != nil {
		return
	}
	var fileLen int
	if fileLen, err = fmt.Fprint(file, os.Getpid()); err != nil {
		return
	}
	if err = file.Truncate(int64(fileLen)); err != nil {
		return
	}
	err = file.Sync()
	return
}

// Remove remove file
func (file *LockFile) Remove() error {
	defer file.Close()

	if err := file.Unlock(); err != nil {
		return err
	}

	return os.Remove(file.Name())
}

// ReadPidFile open pid file
func ReadPidFile(name string) (pid int, err error) {
	var file *os.File
	if file, err = os.OpenFile(name, os.O_RDONLY, 0600); err != nil {
		return
	}
	defer file.Close()

	lock := &LockFile{file}
	pid, err = lock.ReadPid()
	return
}

// ReadPid read pid from pid file
func (file *LockFile) ReadPid() (pid int, err error) {
	if _, err = file.Seek(0, os.SEEK_SET); err != nil {
		return
	}
	_, err = fmt.Fscan(file, &pid)
	return
}
