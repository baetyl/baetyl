package utils

import (
	"compress/flate"

	"github.com/mholt/archiver"
)

var defaultTarGz = &archiver.TarGz{
	CompressionLevel: flate.DefaultCompression,
	Tar: &archiver.Tar{
		MkdirAll:          true,
		OverwriteExisting: true,
	},
}

// TarGz tarGz source files to destination file
func TarGz(sources []string, destination string) error {
	return defaultTarGz.Archive(sources, destination)
}

// UntarGz untarGz source file to destination
func UntarGz(source, destination string) error {
	return defaultTarGz.Unarchive(source, destination)
}
