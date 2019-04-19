package cmd

import (
	"fmt"
	"os"
	"syscall"

	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/fsnotify/fsnotify"
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
	fmt.Fprintln(os.Stdout, "openedge stopping...")
	watcher()
	fmt.Fprintln(os.Stdout, "openedge stopped")
	return nil
}

func watcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Errorf("error:", err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					done <- true
					break
				}
			case err := <-watcher.Errors:
				fmt.Errorf("error:", err)
			}
		}
	}()

	err = watcher.Add(openedge.DefaultPidFile)
	if err != nil {
		fmt.Errorf("error:", err)
	}
	<-done
}
