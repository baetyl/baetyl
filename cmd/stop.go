package cmd

import (
    "log"
    "syscall"

    "github.com/baidu/openedge/daemon"
    "github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
    Use:   "stop",
    Short: "stop openedge",
    Long:  ``,
    Run: func(cmd *cobra.Command, args []string) {
        stop()
    },
}

func init() {
    rootCmd.AddCommand(stopCmd)
}

func stop() {
    cntxt := &daemon.Context{
        PidFileName: "/var/run/openedge.pid",
    }

    d, err := cntxt.Search()
    if err != nil {
        log.Fatalln("Unable send signal to the daemon: ", err)
        return
    }
    err = d.Signal(syscall.SIGTERM)
    if err != nil {
        log.Fatalln("Failed to stop openedge:", err)
        return
    }
    log.Println("openedge stoped.")
    return
}
