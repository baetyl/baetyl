package daemon

import (
    "errors"
    "os"
)

var errNotSupported = errors.New("deamon: Non-POSIX OS  is not supported")

const (
    MARK_NAME  = "_OPENEDGE_DAEMON"
    MARK_VALUE = "1"
    FILE_PERM  = os.FileMode(0640)
)

// WasReborn was reborn
func WasReborn() bool {
    return os.Getenv(MARK_NAME) == MARK_VALUE
}

// Reborn reborn
func (d *Context) Reborn() (child *os.Process, errNotSupported error) {
    return d.reborn()
}

// Release release
func (d *Context) Release() (err error) {
    return d.release()
}

func (d *Context) Search() (daemon *os.Process, err error) {
    return d.search()
}
