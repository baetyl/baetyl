package module

import (
	"flag"
	"os"
	"path/filepath"
)

// Flags command-line flags
type Flags struct {
	WorkDir string
	Config  string
	Help    bool
}

// ParseFlags parses the command-line flags
func ParseFlags() (*Flags, error) {
	f := new(Flags)
	cwd, err := os.Executable()
	if err != nil {
		return nil, err
	}
	cwd, err = filepath.EvalSymlinks(cwd)
	if err != nil {
		return nil, err
	}
	//  default
	f.WorkDir = filepath.Dir(filepath.Dir(cwd))
	f.Config = filepath.Join("conf", "conf.yml")
	flag.StringVar(
		&f.WorkDir,
		"w",
		f.WorkDir,
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
	f.WorkDir, err = filepath.Abs(f.WorkDir)
	if err != nil {
		return nil, err
	}
	err = os.Chdir(f.WorkDir)
	if err != nil {
		return nil, err
	}
	return f, nil
}
