package utils

import (
	"fmt"
	"os"
	"path"
	"strings"
)

// DirExists checkes dir exists
func DirExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return fi.IsDir()
}

// FileExists checkes file exists
func FileExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return !fi.IsDir()
}

// ParseVolumes parses valume mapping of docker container
func ParseVolumes(root string, volumes []string) ([]string, error) {
	result := []string{}
	for _, v := range volumes {
		if strings.Contains(v, "..") {
			return nil, fmt.Errorf("volume (%s) contains invalid string (..)", v)
		}
		parts := strings.Split(v, ":")
		if len(parts) == 0 {
			continue
		}
		for index := 0; index < len(parts); index++ {
			parts[index] = path.Clean(parts[index])
		}
		if len(parts) == 1 {
			parts = append(parts, parts[0])
		}
		if !path.IsAbs(parts[1]) {
			return nil, fmt.Errorf("volume (%s) in container is not absolute", v)
		}
		parts[0] = path.Join(root, parts[0])
		result = append(result, strings.Join(parts, ":"))
	}
	return result, nil
}
