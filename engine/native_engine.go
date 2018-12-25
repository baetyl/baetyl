package engine

import (
	"os"
	"os/exec"
	"path"
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
	log     *logger.Entry
}

// NewNativeEngine create a new native engine
func NewNativeEngine(context *Context) Inner {
	return &NativeEngine{
		context: context,
		log:     logger.WithFields("mode", "native"),
	}
}

// Prepare dummy
func (e *NativeEngine) Prepare(_ string) error {
	return nil
}

// Create creates a new native process
func (e *NativeEngine) Create(m config.Module) (Worker, error) {
	args := []string{m.UniqueName()}
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
		args = append(args, "-c", path.Join(e.context.PWD, "var", "db", "openedge", "module", m.Name, "module.yml"))
	} else {
		args = append(args, m.Params...)
	}
	e.log.Debugln(m.Entry, args)
	return NewNativeProcess(&NativeSpec{
		module:  &m,
		context: e.context,
		exec:    m.Entry,
		argv:    args,
		attr: os.ProcAttr{
			Dir: e.context.PWD,
			Env: utils.AppendEnv(m.Env, true),
			Files: []*os.File{
				os.Stdin,
				os.Stdout,
				os.Stderr,
			},
		},
	}), nil
}
