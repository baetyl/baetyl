package cmd

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/baidu/openedge/daemon"
	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/master"
	"github.com/baidu/openedge/sdk-go/openedge"
	"github.com/baidu/openedge/utils"
	"github.com/spf13/cobra"
)

// compile variables
var (
	workDir  string
	confPath string
)

const defaultConfig = "etc/openedge/openedge.yml"

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start openedge on background",
	Long:  ``,
	Run:   execute,
}

func init() {
	startCmd.Flags().StringVarP(&workDir, "workdir", "w", "", "workdir")
	startCmd.Flags().StringVarP(&confPath, "config", "c", "", "config path")
	rootCmd.AddCommand(startCmd)
}

func execute(cmd *cobra.Command, args []string) {
	workDir, confPath = workPath(workDir, confPath)
	cfg, err := readConfig(workDir, confPath)
	if err != nil {
		logger.Fatalln("failed to read configuration.")
		return
	}
	if runtime.GOOS == "windows" {
		start(workDir, cfg)
	} else {
		onDaemon(workDir, cfg)
	}
}

func onDaemon(workDir string, cfg *master.Config) {
	cntxt := &daemon.Context{
		PidFileName: "/var/run/openedge.pid",
		PidFilePerm: 0644,
		Umask:       027,
	}

	args := []string{"openedge", "start"}
	if workDir != "" {
		args = append(args, "-w", workDir)
	}

	if confPath != "" {
		args = append(args, "-c", confPath)
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

	start(workDir, cfg)
}

func start(pwd string, cfg *master.Config) {
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

func workPath(workDir string, confPath string) (string, string) {
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
	if confPath == "" {
		confPath = defaultConfig
	}
	return workDir, confPath
}

// TODO: make it utils
func readConfig(pwd, cfgFile string) (*master.Config, error) {
	var cfg master.Config
	err := utils.LoadYAML(path.Join(pwd, cfgFile), &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load %s: %s", cfgFile, err.Error())
	}
	err = defaults(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to set default config: %s", err.Error())
	}
	return &cfg, err
}

func defaults(c *master.Config) error {
	if runtime.GOOS == "linux" {
		err := os.MkdirAll("/var/run", os.ModePerm)
		if err != nil {
			logger.WithError(err).Errorf("failed to make dir: /var/run")
		}
		c.Server.Address = "unix:///var/run/openedge.sock"
		utils.SetEnv(openedge.EnvMasterAPIKey, c.Server.Address)
	} else {
		if c.Server.Address == "" {
			c.Server.Address = "tcp://127.0.0.1:50050"
		}
		addr := c.Server.Address
		uri, err := url.Parse(addr)
		if err != nil {
			return err
		}
		if c.Mode == "docker" {
			parts := strings.SplitN(uri.Host, ":", 2)
			addr = fmt.Sprintf("tcp://host.docker.internal:%s", parts[1])
		}
		utils.SetEnv(openedge.EnvMasterAPIKey, addr)
	}
	utils.SetEnv(openedge.EnvHostOSKey, runtime.GOOS)
	utils.SetEnv(openedge.EnvRunningModeKey, c.Mode)
	return nil
}
