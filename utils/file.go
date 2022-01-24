package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/utils"
)

func CreateWriteFile(filePath string, data []byte) error {
	dir := filepath.Dir(filePath)
	if !utils.DirExists(dir) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errors.Trace(err)
		}
	}
	if err := ioutil.WriteFile(filePath, data, 0755); err != nil {
		return errors.Trace(err)
	}
	return nil
}
