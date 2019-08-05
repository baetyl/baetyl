package utils

import (
	"crypto/md5"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"strings"
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

// EscapeIntercept prevent path escape
func EscapeIntercept(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", nil
	}
	basePath := ""
	if path[0] == '/' {
		abs, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return "", err
		}
		basePath = abs
	}
	rel, err := filepath.Rel(basePath, path)
	if err != nil {
		return "", err
	}
	n := len(rel)
	out := make([]byte, n)
	p, q := 0, 0 // p: path index , q: out index
	for p < n {
		switch {
		case os.IsPathSeparator(rel[p]):
			out[q] = rel[p]
			p++
			q++
		case rel[p] == '.' && (p+1 == n || os.IsPathSeparator(rel[p+1])):
			// . element
			out[q] = rel[p]
			p++
			q++
		case rel[p] == '.' && rel[p+1] == '.':
			// .. element
			if p+2 == n {
				out[q] = rel[p]
				q++
			} else if os.IsPathSeparator(rel[p+2]) {
				p++
			}
			p += 2
		default:
			for ; p < n && !os.IsPathSeparator(rel[p]); p++ {
				out[q] = rel[p]
				q++
			}
		}
	}
	return string(out[0:q]), nil
}
