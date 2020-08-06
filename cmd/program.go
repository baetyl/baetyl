package cmd

import (
	"fmt"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/kardianos/service"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
)

func init() {
	rootCmd.AddCommand(programCmd)
}

var programCmd = &cobra.Command{
	Use:   "program",
	Short: "Control a program by Baetyl",
	Long:  `Loads program's information from program.yml, then starts and waits the program to stop.`,
	Run: func(_ *cobra.Command, _ []string) {
		run()
	},
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

	Stderr string `yaml:"stderr" json:"stderr"`
	Stdout string `yaml:"stdout" json:"stdout"`
}

var logger service.Logger

type program struct {
	cfg  Config
	cmd  *exec.Cmd
	svc  service.Service
	exit chan struct{}
}

func (p *program) Start(s service.Service) error {
	// Look for exec.
	// Verify home directory.
	fullExec, err := exec.LookPath(p.cfg.Exec)
	if err != nil {
		return fmt.Errorf("Failed to find executable %q: %v", p.cfg.Exec, err)
	}

	p.cmd = exec.Command(fullExec, p.cfg.Args...)
	p.cmd.Dir = p.cfg.Dir
	p.cmd.Env = append(os.Environ(), p.cfg.Env...)

	go p.run()
	return nil
}
func (p *program) run() {
	logger.Info("Starting ", p.cfg.DisplayName)
	defer func() {
		if service.Interactive() {
			p.Stop(p.svc)
		} else {
			p.svc.Stop()
		}
	}()

	if p.cfg.Stderr != "" {
		f, err := os.OpenFile(p.cfg.Stderr, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
		if err != nil {
			logger.Warningf("Failed to open std err %q: %v", p.cfg.Stderr, err)
			return
		}
		defer f.Close()
		p.cmd.Stderr = f
	}
	if p.cfg.Stdout != "" {
		f, err := os.OpenFile(p.cfg.Stdout, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
		if err != nil {
			logger.Warningf("Failed to open std out %q: %v", p.cfg.Stdout, err)
			return
		}
		defer f.Close()
		p.cmd.Stdout = f
	}

	err := p.cmd.Run()
	if err != nil {
		logger.Warningf("Error running: %v", err)
	}

	return
}
func (p *program) Stop(s service.Service) error {
	close(p.exit)
	logger.Info("Stopping ", p.cfg.DisplayName)
	if p.cmd.Process != nil {
		p.cmd.Process.Kill()
	}
	if service.Interactive() {
		os.Exit(0)
	}
	return nil
}

func run() {
	prg := &program{
		exit: make(chan struct{}),
	}
	err := utils.LoadYAML("program.yml", &prg.cfg)
	if err != nil {
		log.Fatal(err)
	}

	svcCfg := &service.Config{
		Name:        prg.cfg.Name,
		DisplayName: prg.cfg.DisplayName,
		Description: prg.cfg.Description,
	}

	prg.svc, err = service.New(prg, svcCfg)
	if err != nil {
		log.Fatal(err)
	}

	errs := make(chan error, 5)
	logger, err = prg.svc.Logger(errs)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			err := <-errs
			if err != nil {
				log.Print(err)
			}
		}
	}()

	err = prg.svc.Run()
	if err != nil {
		logger.Error(err)
	}
}
