package cmd

import (
	"fmt"
	"os"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/kardianos/service"
	"github.com/spf13/cobra"

	"github.com/baetyl/baetyl/v2/ami"
	"github.com/baetyl/baetyl/v2/config"
)

func init() {
	rootCmd.AddCommand(serviceCmd)
	serviceCmd.AddCommand(startCmd)
	serviceCmd.AddCommand(stopCmd)
	serviceCmd.AddCommand(restartCmd)
	restartCmd.AddCommand(restartAllCmd)
	serviceCmd.AddCommand(listCmd)
}

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Control all services of Baetyl",
	Long:  `Baetyl can control all services by following commands: start, stop, restart.`,
	Args:  cobra.MinimumNArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		action := args[0]
		name := args[1]
		cfg := &service.Config{
			Name: name,
		}
		svc, err := service.New(nil, cfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		err = service.Control(svc, action)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the service by service name",
	Args:  checkServiceName(),
	Run: func(_ *cobra.Command, args []string) {
		if err := newService(args).Start(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the service by service name",
	Args:  checkServiceName(),
	Run: func(_ *cobra.Command, args []string) {
		if err := newService(args).Stop(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the service by service name",
	Args:  checkServiceName(),
	Run: func(_ *cobra.Command, args []string) {
		if err := newService(args).Restart(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}

var restartAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Restart all baetyl services",
	Args:  cobra.NoArgs,
	Run: func(_ *cobra.Command, args []string) {
		restartAll()
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all baetyl services",
	Args:  cobra.NoArgs,
	Run: func(_ *cobra.Command, args []string) {
		listAll()
	},
}

func checkServiceName() cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("requires service name, please input service name")
		}
		return nil
	}
}

func newService(args []string) service.Service {
	cfg := &service.Config{
		Name: args[0],
	}
	svc, err := service.New(nil, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return svc
}

func restartAll() {
	var err error
	var l *log.Logger
	l, err = log.Init(log.Config{Level: "debug", Encoding: "console"})
	if err != nil {
		log.L().Error("failed to init logger", log.Error(err))
		return
	}
	defer func() {
		if err != nil {
			l.Error(err.Error())
		}
	}()

	amiConfig := config.AmiConfig{}
	err = utils.SetDefaults(&amiConfig)
	if err != nil {
		return
	}
	am, err := ami.NewAMI(mode, amiConfig)
	if err != nil {
		return
	}
	var allstats []specv1.AppStats
	sysstats, _ := am.StatsApps(context.EdgeSystemNamespace())
	userstats, _ := am.StatsApps(context.EdgeNamespace())
	allstats = append(allstats, sysstats...)
	allstats = append(allstats, userstats...)
	log.L().Info("stats apps", log.Any("all", allstats), log.Error(err))
	for _, appstats := range allstats {
		for _, ins := range appstats.InstanceStats {
			if ins.PPid == 1 {
				err = restartSvc(ins.Name)
				if err != nil {
					log.L().Error("failed to restart svc", log.Any("name", ins.Name), log.Error(err))
				} else {
					log.L().Info("restart svc success", log.Any("name", ins.Name))
				}
			}
		}
	}
}

func restartSvc(svcName string) error {
	cfg := &service.Config{
		Name: svcName,
	}
	svc, err := service.New(nil, cfg)
	if err != nil {
		return err
	}
	return svc.Restart()
}

func listAll() {
	var err error
	var l *log.Logger
	l, err = log.Init(log.Config{Level: "debug", Encoding: "console"})
	if err != nil {
		log.L().Error("failed to init logger", log.Error(err))
		return
	}
	defer func() {
		if err != nil {
			l.Error(err.Error())
		}
	}()

	amiConfig := config.AmiConfig{}
	err = utils.SetDefaults(&amiConfig)
	if err != nil {
		return
	}
	am, err := ami.NewAMI(mode, amiConfig)
	if err != nil {
		return
	}
	var allstats []specv1.AppStats
	sysstats, _ := am.StatsApps(context.EdgeSystemNamespace())
	userstats, _ := am.StatsApps(context.EdgeNamespace())
	allstats = append(allstats, sysstats...)
	allstats = append(allstats, userstats...)
	for _, appstats := range allstats {
		for _, ins := range appstats.InstanceStats {
			if ins.PPid == 1 {
				log.L().Info("listing svc", log.Any("name", ins.Name))
			}
		}
	}
	log.L().Info("listing over")
}
