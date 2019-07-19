package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"

	"github.com/baidu/openedge/master"
	"github.com/baidu/openedge/utils"
	"github.com/spf13/cobra"
)

const defaultConfFile = "etc/openedge/openedge.yml"

// compile variables
var (
	workDir  string
	confFile string
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start openedge on background",
	Long:  ``,
	Run:   start,
}

func init() {
	startCmd.Flags().StringVarP(&workDir, "workdir", "w", "", "work directory of openedge")
	startCmd.Flags().StringVarP(&confFile, "config", "c", "", "config path of openedge")
	rootCmd.AddCommand(startCmd)
}

func start(cmd *cobra.Command, args []string) {
	err := startInternal()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func startInternal() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable: %s", err.Error())
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("failed to get path of executable: %s", err.Error())
	}
	if workDir == "" {
		workDir = path.Dir(path.Dir(exe))
	}
	workDir, err = filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of work directory: %s", err.Error())
	}
	err = os.Chdir(workDir)
	if err != nil {
		return fmt.Errorf("failed to change work directory: %s", err.Error())
	}
	if confFile == "" {
		confFile = defaultConfFile
	}
	var cfg master.Config
	err = utils.LoadYAML(path.Join(workDir, confFile), &cfg)
	if err != nil {
		return fmt.Errorf("failed to load config: %s", err.Error())
	}

	m, err := master.New(workDir, cfg, Version)
	if err != nil {
		return fmt.Errorf("failed to create master: %s", err.Error())
	}
	defer m.Close()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	signal.Ignore(syscall.SIGPIPE)
	<-sig
	return nil
}
