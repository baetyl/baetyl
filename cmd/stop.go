package cmd

import (
	"fmt"
	"os"
	"syscall"

	openedge "github.com/baidu/openedge/sdk/openedge-go"
	daemon "github.com/sevlyar/go-daemon"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop openedge and all services",
	Long:  ``,
	Run:   stop,
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func stop(cmd *cobra.Command, args []string) {
	err := stopInternal()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func stopInternal() error {
	cntxt := &daemon.Context{
		PidFileName: openedge.DefaultPidFile,
	}
	d, err := cntxt.Search()
	if err != nil {
		return fmt.Errorf("failed to search openedge: %s", err.Error())
	}
	err = d.Signal(syscall.SIGTERM)
	if err != nil {
		return fmt.Errorf("failed to stop openedge: %s", err.Error())
	}
	return nil
}
