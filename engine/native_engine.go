package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/baidu/openedge/module/config"
	"github.com/baidu/openedge/module/logger"
	"github.com/baidu/openedge/module/utils"
)

// NativeEngine native engine
type NativeEngine struct {
	context *Context
	pwd     string
	log     *logger.Entry
}

// NewNativeEngine create a new native engine
func NewNativeEngine(context *Context) (Inner, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return &NativeEngine{
		context: context,
		pwd:     pwd,
		log:     logger.WithFields("mode", "native"),
	}, nil
}

// Prepare dummy
func (e *NativeEngine) Prepare(_ string) error {
	return nil
}

// Create creates a new native process
func (e *NativeEngine) Create(m config.Module) (Worker, error) {
	args := []string{m.Name}
	if runtime.GOOS == "windows" {
		if !strings.Contains(filepath.Base(m.Entry), ".") {
			m.Entry = m.Entry + ".exe"
		} else if strings.HasSuffix(m.Entry, ".py") {
			prog, err := exec.LookPath("python.exe")
			if err != nil {
				return nil, err
			}
			args = append(args, m.Entry)
			m.Entry = prog
		} else if strings.HasSuffix(m.Entry, ".js") {
			prog, err := exec.LookPath("node.exe")
			if err != nil {
				return nil, err
			}
			args = append(args, m.Entry)
			m.Entry = prog
		}
	}
	if len(m.Params) == 0 {
		args = append(args, "-c", fmt.Sprintf("app/%s/conf.yml", m.Mark))
	} else {
		args = append(args, m.Params...)
	}
	e.log.Debugln(m.Entry, args)
	return NewNativeProcess(&NativeSpec{
		Spec: Spec{
			Name:    m.Name,
			Restart: m.Restart,
			Grace:   e.context.Grace,
			Logger:  e.log.WithFields("module", m.Name),
		},
		Exec: m.Entry,
		Argv: args,
		Attr: os.ProcAttr{
			Dir: e.pwd,
			Env: utils.AppendEnv(m.Env, true),
			Files: []*os.File{
				os.Stdin,
				os.Stdout,
				os.Stderr,
			},
		},
	}), nil
}
