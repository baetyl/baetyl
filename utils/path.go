package utils

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

// PathExists checks path exists
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// DirExists checks dir exists
func DirExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return fi.IsDir()
}

// FileExists checks file exists
func FileExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return !fi.IsDir()
}

// WriteFile writes data into file in chunk mode
func WriteFile(fn string, r io.Reader) error {
	f, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	return err
}

// CopyFile copy data from one file to another
func CopyFile(s, t string) error {
	sf, err := os.Open(s)
	if err != nil {
		return err
	}
	defer sf.Close()

	return WriteFile(t, sf)
}

// CalculateFileMD5 calculates file MD5
func CalculateFileMD5(fn string) (string, error) {
	f, err := os.Open(fn)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hasher := md5.New()
	_, err = io.Copy(hasher, f)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(hasher.Sum(nil)), nil
}

// CreateSymlink create symlink of target
func CreateSymlink(target, symlink string) error {
	if PathExists(symlink) {
		return nil
	}
	err := os.Symlink(target, symlink)
	if err != nil {
		return fmt.Errorf("failed to make symlink %s of %s: %s", target, symlink, err.Error())
	}
	return nil
}
