package cmd

import (
	"fmt"
	"os"

	openedge "github.com/baidu/openedge/sdk/openedge-go"
	daemon "github.com/sevlyar/go-daemon"
	"github.com/shirou/gopsutil/process"
	"github.com/spf13/cobra"
)

// restartCmd represents the restart command
var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "restart openedge",
	Long:  ``,
	Run:   restart,
}

func restart(cmd *cobra.Command, args []string) {
	err := restartInternal()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func restartInternal() error {
	pid, err := daemon.ReadPidFile(openedge.DefaultPidFile)
	if err != nil {
		err = fmt.Errorf("failed to read existed pid file: %s", err.Error())
		return err
	}
	process := &process.Process{Pid: int32(pid)}
	_, err = process.Status()
	if err != nil {
		return fmt.Errorf("openedge didn't start, please start openedge first: %s", err.Error())
	}
	fmt.Fprintln(os.Stdout, "restart openedge")
	err = stopInternal()
	if err != nil {
		return fmt.Errorf("failed to stop openedge: %s", err.Error())
	}
	fmt.Fprintln(os.Stdout, "openedge restarting...")
	err = startInternal()
	if err != nil {
		return fmt.Errorf("failed to start openedge: %s", err.Error())
	}
	fmt.Fprintln(os.Stdout, "openedge restarted")
	return nil
}

func init() {
	restartCmd.Flags().StringVarP(&workDir, "workdir", "w", "", "work directory of openedge")
	restartCmd.Flags().StringVarP(&confFile, "config", "c", "", "config path of openedge")
	rootCmd.AddCommand(restartCmd)
}
