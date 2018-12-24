package module

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// Flags command-line flags
type Flags struct {
	Config string
	Help   bool
}

// ParseFlags parses the command-line flags
func ParseFlags(defaultConfigPath string) (*Flags, error) {
	f := new(Flags)
	cwd, err := os.Executable()
	if err != nil {
		return nil, err
	}
	cwd, err = filepath.EvalSymlinks(cwd)
	if err != nil {
		return nil, err
	}
	workDir := filepath.Dir(filepath.Dir(cwd))
	if defaultConfigPath == "" {
		f.Config = filepath.Join("etc", "openedge", "module.yml")
	} else {
		f.Config = defaultConfigPath
	}
	flag.StringVar(
		&workDir,
		"w",
		workDir,
		"working directory",
	)
	flag.StringVar(
		&f.Config,
		"c",
		f.Config,
		"config file path",
	)
	flag.BoolVar(
		&f.Help,
		"h",
		false,
		"show this help",
	)
	flag.Parse()
	workDir, err = filepath.Abs(workDir)
	if err != nil {
		return nil, err
	}
	err = os.Chdir(workDir)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// PrintUsage prints usage
func PrintUsage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Version of %s: %s\n", os.Args[0], Version)
	flag.Usage()
}
