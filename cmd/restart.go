package cmd

import (
    "log"
    "runtime"
    "time"

    "github.com/spf13/cobra"
)

// restartCmd represents the restart command
var restartCmd = &cobra.Command{
    Use:   "restart",
    Short: "restart openedge",
    Long:  ``,
    Run:   exeRestart,
}

func init() {
    restartCmd.Flags().StringVarP(&workDir, "workdir", "w", "", "workdir")
    restartCmd.Flags().StringVarP(&confPath, "config", "c", "", "config path")
    rootCmd.AddCommand(restartCmd)
}

func exeRestart(cmd *cobra.Command, args []string) {
    if runtime.GOOS == "windows" {
        log.Fatalln("The stop command is temporarily not supported on the Windows platform.")
        return
    }
    stop()
    // TODO wait the previous process release resources
    time.Sleep(2 * time.Second)
    err := start()
    if err != nil {
        log.Fatalln("Restart openedge failed", err)
    }
    log.Println("openedge restarted.")
}
