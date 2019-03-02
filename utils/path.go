package utils

import (
	"crypto/md5"
	"encoding/base64"
	"io"
	"os"
)

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
