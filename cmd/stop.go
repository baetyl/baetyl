package cmd

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
	"time"

	openedge "github.com/baidu/openedge/sdk/openedge-go"
	daemon "github.com/sevlyar/go-daemon"
	"github.com/shirou/gopsutil/process"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop openedge and all services",
	Long:  ``,
	Run:   stop,
}

var timeout string

func init() {
	stopCmd.Flags().StringVarP(&timeout, "timeout", "t", "", "set the timeout for the stop command, unit: s")
	rootCmd.AddCommand(stopCmd)
}

func stop(cmd *cobra.Command, args []string) {
	err := stopInternal()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func stopInternal() error {
	pid, err := daemon.ReadPidFile(openedge.DefaultPidFile)
	if err != nil {
		err = fmt.Errorf("failed to read existed pid file: %s", err.Error())
		return err
	}
	process := &process.Process{Pid: int32(pid)}
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
	fmt.Fprintln(os.Stdout, "openedge stopping...")
	if timeout == "" {
		timeout = "0"
	}
	val, err := strconv.Atoi(timeout)
	if err != nil {
		return fmt.Errorf("timeout should be an integer: %s", err.Error())
	}
	timeout := time.After(time.Second * time.Duration(val))
	for {
		select {
		case <-timeout:
			syscall.Kill(pid, syscall.SIGKILL)
			fmt.Fprintln(os.Stdout, "openedge stopped")
			return nil
		default:
			_, err := process.Status()
			if err != nil {
				fmt.Fprintln(os.Stdout, "openedge stopped")
				return nil
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
}
