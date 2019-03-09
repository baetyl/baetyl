package cmd

import (
    "log"
    "syscall"

    daemon "github.com/sevlyar/go-daemon"
    "github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
    Use:   "stop",
    Short: "stop openedge",
    Long:  ``,
    Run:   stop,
}

func init() {
    rootCmd.AddCommand(stopCmd)
}

func stop(cmd *cobra.Command, args []string) {
    stopInternal()
}

func stopInternal() {
    cntxt := &daemon.Context{
        PidFileName: pidFilePath,
    }
    d, err := cntxt.Search()
    if err != nil {
        log.Fatalln("Unable send signal to the daemon:", err)
        return
    }
    err = d.Signal(syscall.SIGTERM)
    if err != nil {
        log.Fatalln("Failed to stop openedge:", err)
        return
    }

    return
}
