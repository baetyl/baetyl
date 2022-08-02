package program

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/kardianos/service"
)

type Program struct {
	cfg  Config
	cmd  *exec.Cmd
	svc  service.Service
	exit chan struct{}
	log  io.Writer
}

func (p *Program) Start(s service.Service) error {
	// Look for exec.
	// Verify home directory.
	fullExec, err := exec.LookPath(p.cfg.Exec)
	if err != nil {
		return errors.Trace(err)
	}

	if runtime.GOOS == "windows" && strings.Contains(fullExec, ".py") {
		p.cmd = exec.Command("python", fullExec)
	} else if runtime.GOOS == "windows" && strings.Contains(fullExec, ".js") {
		p.cmd = exec.Command("node", fullExec)
	} else {
		p.cmd = exec.Command(fullExec, p.cfg.Args...)
		p.cmd.Dir = p.cfg.Dir
	}
	p.cmd.Env = append(os.Environ(), p.cfg.Env...)
	p.cmd.Stderr = p.log
	p.cmd.Stdout = p.log

	go p.run()
	return nil
}
func (p *Program) run() {
	fmt.Fprintln(p.log, "Program starting", p.cfg.DisplayName)
	defer p.Stop(p.svc)

	err := p.cmd.Run()
	if err != nil {
		fmt.Fprintln(p.log, "Program error:", err)
	}
	return
}
func (p *Program) Stop(s service.Service) error {
	close(p.exit)

	fmt.Fprintln(p.log, "Program stopping", p.cfg.DisplayName)
	if p.cmd.Process != nil {
		p.cmd.Process.Kill()
	}

	os.Exit(0)
	return nil
}
