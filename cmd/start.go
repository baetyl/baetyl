package cmd

import (
	"fmt"
	"os"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/master"
	"github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/baetyl/baetyl/utils"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start baetyl",
	Long:  ``,
	Run:   start,
}

func init() {
	startCmd.Flags().StringVarP(&workDir, "workdir", "w", "", "work directory of baetyl")
	startCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "config path of baetyl")
	rootCmd.AddCommand(startCmd)
}

func start(cmd *cobra.Command, args []string) {
	err := startInternal()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func startInternal() error {
	cfg, err := checkInternal()
	log := logger.InitLogger(cfg.Logger, "baetyl", "master")
	isOTA := utils.FileExists(cfg.OTALog.Path)
	if isOTA {
		log = logger.New(cfg.OTALog, "type", baetyl.OTAMST)
	}
	if err != nil {
		log.WithField(baetyl.OTAKeyStep, baetyl.OTARollingBack).WithError(err).Errorf("failed to start master")
		rberr := master.RollBackMST()
		if rberr != nil {
			log.WithField(baetyl.OTAKeyStep, baetyl.OTAFailure).WithError(rberr).Errorf("failed to roll back")
			return fmt.Errorf("failed to start master: %s; failed to roll back: %s", err.Error(), rberr.Error())
		}
		log.WithField(baetyl.OTAKeyStep, baetyl.OTARolledBack).Infof("master is rolled back")
		return fmt.Errorf("failed to start master: %s", err.Error())
	}

	m, err := master.New(workDir, *cfg, Version, Revision)
	if err != nil {
		log.WithField(baetyl.OTAKeyStep, baetyl.OTARollingBack).WithError(err).Errorf("failed to start master")
		rberr := master.RollBackMST()
		if rberr != nil {
			log.WithField(baetyl.OTAKeyStep, baetyl.OTAFailure).WithError(rberr).Errorf("failed to roll back")
			return fmt.Errorf("failed to start master: %s; failed to roll back: %s", err.Error(), rberr.Error())
		}
		log.WithField(baetyl.OTAKeyStep, baetyl.OTARestarting).Infof("master is restarting")
		return fmt.Errorf("failed to start master: %s", err.Error())
	}
	defer m.Close()
	if master.CommitMST() {
		log.WithField(baetyl.OTAKeyStep, baetyl.OTAUpdated).Infof("master is updated")
	} else if isOTA {
		log.WithField(baetyl.OTAKeyStep, baetyl.OTARolledBack).Infof("master is rolled back")
	}
	return m.Wait()
}
