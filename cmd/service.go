package cmd

import (
	"fmt"
	"os"

	"github.com/kardianos/service"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serviceCmd)
	serviceCmd.AddCommand(startCmd)
	serviceCmd.AddCommand(stopCmd)
	serviceCmd.AddCommand(restartCmd)
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
			Name: name + ".baetyl-edge",
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
		Name: args[0] + ".baetyl-edge",
	}
	svc, err := service.New(nil, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return svc
}
