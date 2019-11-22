package utils

import (
	"github.com/mholt/archiver"
)

var defaultTar = &archiver.Tar{
	MkdirAll:          true,
	OverwriteExisting: true,
}

// Tar tar source files to destination file
func Tar(sources []string, destination string) error {
	return defaultTar.Archive(sources, destination)
}

// Untar untar source file to destination
func Untar(source, destination string) error {
	return defaultTar.Unarchive(source, destination)
}
