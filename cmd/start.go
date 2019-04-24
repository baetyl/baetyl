package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"

	"github.com/baidu/openedge/master"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/baidu/openedge/utils"
	daemon "github.com/sevlyar/go-daemon"
	"github.com/shirou/gopsutil/process"
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

	args := []string{"openedge", "start"}
	if workDir != "" {
		args = append(args, "-w", workDir)
	}

	if confFile != "" {
		args = append(args, "-c", confFile)
	}

	ctx := &daemon.Context{
		PidFileName: openedge.DefaultPidFile,
		PidFilePerm: 0644,
		Umask:       027,
		Args:        args,
	}
	if cfg.Logger.Path != "" {
		os.MkdirAll(path.Dir(cfg.Logger.Path), 0644)
		ctx.LogFileName = cfg.Logger.Path + ".console"
	} else {
		ctx.LogFileName = "openedge.log.console"
	}

	if utils.FileExists(openedge.DefaultPidFile) && !daemon.WasReborn() {
		pid, err := daemon.ReadPidFile(openedge.DefaultPidFile)
		if err != nil {
			err = fmt.Errorf("Failed to read existed pid file: %s", err.Error())
			return err
		}
		process := &process.Process{Pid: int32(pid)}
		_, err = process.Status()
		if err != nil {
			os.Remove(openedge.DefaultPidFile)
		}
	}

	d, err := ctx.Reborn()
	if err != nil {
		if err == daemon.ErrWouldBlock {
			err = fmt.Errorf("OpenEdge has been started, please do not start again")
		}
		return err
	}

	if d != nil {
		return nil
	}

	if utils.PathExists(openedge.DefaultSockFile) {
		err = os.RemoveAll(openedge.DefaultSockFile)
		if err != nil {
			return fmt.Errorf("failed to remove sock file: %s", err.Error())
		}
	}

	defer ctx.Release()

	m, err := master.New(workDir, cfg, Version)
	if err != nil {
		return fmt.Errorf("failed to create master: %s", err.Error())
	}
	defer m.Close()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	signal.Ignore(syscall.SIGPIPE)
	fmt.Fprintln(os.Stdout, "OpenEdge started")
	<-sig
	return nil
}
