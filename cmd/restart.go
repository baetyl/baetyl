package cmd

import (
    "log"
    "os"
    "syscall"
    "time"

    "github.com/baidu/openedge/daemon"
    "github.com/spf13/cobra"
)

var (
    reWorkDir  string
    reConfPath string
)

// restartCmd represents the restart command
var restartCmd = &cobra.Command{
    Use:   "restart",
    Short: "restart openedge",
    Long:  ``,
    Run:   exeRestart,
}

func init() {
    restartCmd.Flags().StringVarP(&reWorkDir, "workdir", "w", "", "workdir")
    restartCmd.Flags().StringVarP(&reConfPath, "config", "c", "", "config path")
    rootCmd.AddCommand(restartCmd)
}

func exeRestart(cmd *cobra.Command, args []string) {
    stop()
    restart(reWorkDir, reConfPath)
}

func stopDaemon() {
    cntxt := daemon.Context{
        PidFileName: "/var/run/openedge.pid",
    }

    d, err := cntxt.Search()
    if err != nil {
        log.Fatalln("Unable send signal to the daemon:", err)
    }
    err = d.Signal(syscall.SIGTERM)
    if err != nil {
        log.Fatalln("Failed to restart openedge, cannot stop previous openedge:", err)
    }
}

func restart(reWorkDir, reConfPath string) {
    args := []string{"openedge", "start"}
    if reWorkDir != "" {
        args = append(args, "-w", reWorkDir)
    }
    if reConfPath != "" {
        args = append(args, "-c", reConfPath)
    }

    devNull, err := os.Open(os.DevNull)
    if err != nil {
        log.Fatalln("Failed to restart openedge, cannot get null device:", err)
    }
    // TODO wait the previous process release resources
    time.Sleep(2 * time.Second)
    attr := &os.ProcAttr{
        Files: []*os.File{os.Stdin, devNull, os.Stderr, devNull},
    }
    _, err = os.StartProcess(os.Args[0], args, attr)
    if err != nil {
        log.Fatalln("Failed to restart openedge", err)
        return
    }
    log.Println("openedge restarted")
    return
}
