package native

// PackageConfigPath to meta data
const PackageConfigPath = "package.yml"

// Package of native image
type Package struct {
	Entry string `yaml:"entry" json:"entry"`
}
