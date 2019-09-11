package utils

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path"
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
	if !PathExists(symlink) {
		err := os.Symlink(target, symlink)
		if err != nil {
			return fmt.Errorf("failed to make symlink %s of %s: %s", target, symlink, err.Error())
		}
	}
	return nil
}

// CreateCwdSymlink create symlink of target in current work directory
func CreateCwdSymlink(pwd, target, symlink string) error {
	t := path.Join(pwd, target)
	s := path.Join(pwd, symlink)
	err := CreateSymlink(t, s)
	if err != nil {
		return err
	}
	return nil
}

// CreateDirAndSymlink Create Dir and related symlink
func CreateDirAndSymlink(pwd, current, previous string) error {
	var err error
	if !PathExists(path.Join(pwd, previous)) {
		err = os.MkdirAll(path.Join(pwd, current), 0755)
		if err != nil {
			return fmt.Errorf("failed to make directory: %s", err.Error())
		}
		err = CreateCwdSymlink(pwd, current, previous)
		if err != nil {
			return err
		}
	} else {
		err = CreateCwdSymlink(pwd, previous, current)
		if err != nil {
			return err
		}
	}
	return nil
}
