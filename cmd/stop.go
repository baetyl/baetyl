package cmd

import (
	"fmt"
	"math"
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
	stopCmd.Flags().StringVarP(&timeout, "timeout", "t", "", "set the timeout in second for the stop command")
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
	timeout, err := strconv.Atoi(timeout)
	if err != nil {
		return fmt.Errorf("timeout should be an integer: %s", err.Error())
	}
	if timeout <= 0 {
		timeout = math.MaxInt32
	}
	loopTicker := time.NewTicker(200 * time.Millisecond)
	defer loopTicker.Stop()
	timeoutTimer := time.NewTimer(time.Duration(timeout) * time.Second)
	defer timeoutTimer.Stop()
	for {
		select {
		case <-timeoutTimer.C:
			syscall.Kill(pid, syscall.SIGKILL)
			fmt.Fprintln(os.Stdout, "openedge killed since timeout")
			return nil
		case <-loopTicker.C:
			_, err := process.Status()
			if err != nil {
				fmt.Fprintln(os.Stdout, "openedge stopped")
				return nil
			}
		}
	}
}
