package cmd

import (
    "log"

    daemon "github.com/sevlyar/go-daemon"
    "github.com/spf13/cobra"
)

// restartCmd represents the restart command
var restartCmd = &cobra.Command{
    Use:   "restart",
    Short: "restart openedge and all services",
    Long:  ``,
    Run:   restart,
}

func init() {
    restartCmd.Flags().StringVarP(&workDir, "workdir", "w", "", "work directory of openedge")
    restartCmd.Flags().StringVarP(&confFile, "config", "c", "", "config path of openedge")
    rootCmd.AddCommand(restartCmd)
}

func restart(cmd *cobra.Command, args []string) {
    stopInternal()
    cntxt := &daemon.Context{
        PidFileName: pidFilePath,
    }
    _, err := cntxt.Search()
    if err != nil {
        startInternal()
    } else {
        log.Println("openedge stop failed")
    }
    log.Println("openedge restarted.")
}
