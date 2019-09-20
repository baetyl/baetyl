package cmd

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/baetyl/baetyl/master"
	"github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/baetyl/baetyl/utils"
	"github.com/spf13/cobra"
)

const defaultConfFile = "etc/baetyl/conf.yml"

// compile variables
var (
	workDir string
	cfgFile string
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "check baetyl and its configuration",
	Long:  ``,
	Run:   check,
}

func init() {
	checkCmd.Flags().StringVarP(&workDir, "workdir", "w", "", "work directory of baetyl")
	checkCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "config path of baetyl")
	rootCmd.AddCommand(checkCmd)
}

func check(cmd *cobra.Command, args []string) {
	_, err := checkInternal()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	log.Printf(baetyl.CheckOK)
}

func checkInternal() (*master.Config, error) {
	// backward compatibility
	err := addSymlinkCompatible()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	cfg := &master.Config{File: defaultConfFile}
	utils.UnmarshalYAML(nil, cfg) // default config
	exe, err := os.Executable()
	if err != nil {
		return cfg, fmt.Errorf("failed to get executable: %s", err.Error())
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return cfg, fmt.Errorf("failed to get path of executable: %s", err.Error())
	}
	if workDir == "" {
		workDir = path.Dir(path.Dir(exe))
	}
	workDir, err = filepath.Abs(workDir)
	if err != nil {
		return cfg, fmt.Errorf("failed to get absolute path of work directory: %s", err.Error())
	}

	if err = os.Chdir(workDir); err != nil {
		return cfg, fmt.Errorf("failed to change work directory: %s", err.Error())
	}

	if cfgFile != "" {
		cfg.File = cfgFile
	}
	if utils.FileExists(cfg.File) {
		err = utils.LoadYAML(cfg.File, cfg)
		if err != nil {
			return cfg, fmt.Errorf("failed to load config: %s", err.Error())
		}
	} else {
		log.Printf("config file (%s) not found, to use default config", cfg.File)
	}

	if err = cfg.Validate(); err != nil {
		return cfg, fmt.Errorf("config invalid: %s", err.Error())
	}
	return cfg, nil
}

func addSymlinkCompatible() error {
	if utils.PathExists(baetyl.PreviousMasterConfDir) {
		err := utils.CreateSymlink(path.Base(baetyl.PreviousMasterConfDir), baetyl.DefaultMasterConfDir)
		if err != nil {
			return err
		}
		err = utils.CreateSymlink(path.Base(baetyl.PreviousMasterConfFile), baetyl.DefaultMasterConfFile)
		if err != nil {
			return err
		}
	}
	if utils.PathExists(baetyl.PreviousDBDir) {
		err := utils.CreateSymlink(path.Base(baetyl.PreviousDBDir), baetyl.DefaultDBDir)
		if err != nil {
			return err
		}
	}
	return nil
}
