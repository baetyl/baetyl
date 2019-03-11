package cmd

import (
    "log"

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
    startInternal()
    log.Println("openedge restarted.")
}
