package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/master"
	"github.com/baidu/openedge/utils"
	daemon "github.com/sevlyar/go-daemon"
	"github.com/spf13/cobra"
)

// compile variables
var (
	workDir  string
	confFile string
)

const defaultConfFile = "etc/openedge/openedge.yml"
const pidFilePath = "/var/run/openedge.pid"

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
	startInternal()
}

func startInternal() {
	workDir, confFile = workPath()
	cfg, err := readConfig(workDir, confFile)
	if err != nil {
		logger.Fatalln("failed to read configuration.")
		return
	}
	onDaemon(cfg)
}

func onDaemon(cfg *master.Config) {
	cntxt := &daemon.Context{
		PidFileName: pidFilePath,
		PidFilePerm: 0644,
		Umask:       027,
	}

	args := []string{"openedge", "start"}
	if workDir != "" {
		args = append(args, "-w", workDir)
	}

	if confFile != "" {
		args = append(args, "-c", confFile)
	}

	cntxt.Args = args

	defer cntxt.Release()
	d, err := cntxt.Reborn()
	if err != nil {
		log.Fatalln(err)
	}
	if d != nil {
		return
	}

	startMaster(cfg)
}

func startMaster(cfg *master.Config) {
	m, err := master.New(workDir, cfg)
	if err != nil {
		logger.Fatalln("failed to create master:", err.Error())
		return
	}
	defer m.Close()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	signal.Ignore(syscall.SIGPIPE)
	<-sig
}

func workPath() (string, string) {
	exe, err := os.Executable()
	if err != nil {
		logger.Fatalln("failed to get executable path:", err.Error())
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		logger.Fatalln("failed to get realpath of executable:", err.Error())
	}
	if workDir == "" {
		workDir = path.Dir(path.Dir(exe))
	}

	workDir, err = filepath.Abs(workDir)
	if err != nil {
		logger.Fatalln("failed to get absolute path of workdir:", err.Error())
	}
	err = os.Chdir(workDir)
	if err != nil {
		logger.Fatalln("failed to change directory to workdir:", err.Error())
	}
	if confFile == "" {
		confFile = defaultConfFile
	}
	return workDir, confFile
}

// TODO: make it utils
func readConfig(pwd, cfgFile string) (*master.Config, error) {
	var cfg master.Config
	err := utils.LoadYAML(path.Join(pwd, cfgFile), &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load %s: %s", cfgFile, err.Error())
	}
	return &cfg, err
}
