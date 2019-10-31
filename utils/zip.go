package utils

import (
	"compress/flate"
	"github.com/mholt/archiver"
)

var defaultZip = &archiver.Zip{
	CompressionLevel:     flate.DefaultCompression,
	MkdirAll:             true,
	SelectiveCompression: true,
	OverwriteExisting:    true,
}

// Zip zip source files to destination file
func Zip(sources []string, destination string) error {
	return defaultZip.Archive(sources, destination)
}

// Unzip unzip source file to destination
func Unzip(source, destination string) error {
	return defaultZip.Unarchive(source, destination)
}
