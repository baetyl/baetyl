package cmd

import (
	"fmt"
	"os"
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
	timeout := time.After(time.Second * 10)
	finish := make(chan bool)
	go func() {
		for {
			select {
			case <-timeout:
				syscall.Kill(pid, syscall.SIGKILL)
				finish <- true
			default:
				_, err := process.Status()
				if err != nil {
					finish <- true
				}
			}
		}
	}()
	<-finish
	fmt.Fprintln(os.Stdout, "openedge stopped")
	return nil
}
