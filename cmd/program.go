package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/kardianos/service"
	"github.com/spf13/cobra"
	"gopkg.in/natefinch/lumberjack.v2"
)

func init() {
	rootCmd.AddCommand(programCmd)
}

var programCmd = &cobra.Command{
	Use:   "program",
	Short: "Control a program by Baetyl",
	Long:  `Baetyl loads program's information from program.yml, then starts and waits the program to stop.`,
	Run: func(_ *cobra.Command, _ []string) {
		if err := startProgramService(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}

func startProgramService() error {
	prg := &program{
		exit: make(chan struct{}),
		log:  os.Stdout,
	}
	err := utils.LoadYAML("program.yml", &prg.cfg)
	if err != nil {
		return errors.Trace(err)
	}

	if prg.cfg.Logger.Filename != "" {
		prg.log = &lumberjack.Logger{
			Compress:   prg.cfg.Logger.Compress,
			Filename:   prg.cfg.Logger.Filename,
			MaxAge:     prg.cfg.Logger.MaxAge,
			MaxSize:    prg.cfg.Logger.MaxSize,
			MaxBackups: prg.cfg.Logger.MaxBackups,
		}
	}

	svcCfg := &service.Config{
		Name:        prg.cfg.Name,
		DisplayName: prg.cfg.DisplayName,
		Description: prg.cfg.Description,
	}

	prg.svc, err = service.New(prg, svcCfg)
	if err != nil {
		return errors.Trace(err)
	}

	return errors.Trace(prg.svc.Run())
}

// Config is the runner app config structure.
type Config struct {
	Name        string `yaml:"name" json:"name" validate:"nonzero"`
	DisplayName string `yaml:"displayName" json:"displayName"`
	Description string `yaml:"description" json:"description"`

	Dir  string   `yaml:"dir" json:"dir"`
	Exec string   `yaml:"exec" json:"exec"`
	Args []string `yaml:"args" json:"args"`
	Env  []string `yaml:"env" json:"env"`

	Logger log.Config `yaml:"logger" json:"logger"`
}

type program struct {
	cfg  Config
	cmd  *exec.Cmd
	svc  service.Service
	exit chan struct{}
	log  io.Writer
}

func (p *program) Start(s service.Service) error {
	// Look for exec.
	// Verify home directory.
	fullExec, err := exec.LookPath(p.cfg.Exec)
	if err != nil {
		return errors.Trace(err)
	}

	p.cmd = exec.Command(fullExec, p.cfg.Args...)
	p.cmd.Dir = p.cfg.Dir
	p.cmd.Env = append(os.Environ(), p.cfg.Env...)
	p.cmd.Stderr = p.log
	p.cmd.Stdout = p.log

	go p.run()
	return nil
}
func (p *program) run() {
	fmt.Fprintln(p.log, "Program starting", p.cfg.DisplayName)
	defer p.Stop(p.svc)

	err := p.cmd.Run()
	if err != nil {
		fmt.Fprintln(p.log, "Program error:", err)
	}
	return
}
func (p *program) Stop(s service.Service) error {
	close(p.exit)

	fmt.Fprintln(p.log, "Program stopping", p.cfg.DisplayName)
	if p.cmd.Process != nil {
		p.cmd.Process.Kill()
	}

	os.Exit(0)
	return nil
}
