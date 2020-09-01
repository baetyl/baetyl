package program

import (
	"os"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/kardianos/service"
	"gopkg.in/natefinch/lumberjack.v2"
)

const ProgramBinPath = "bin"
const ProgramEntryYaml = "program.yml" // in program package to specify entry
const ProgramServiceYaml = "service.yml"

func Run() error {
	prg := &Program{
		exit: make(chan struct{}),
		log:  os.Stdout,
	}
	err := utils.LoadYAML(ProgramServiceYaml, &prg.cfg)
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
