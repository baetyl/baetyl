package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// restartCmd represents the restart command
var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "restart openedge",
	Long:  ``,
	Run:   restart,
}

func restart(cmd *cobra.Command, args []string) {
	fmt.Fprintln(os.Stdout, "restart openedge")
	err := stopInternal()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	fmt.Fprintln(os.Stdout, "openedge restarting...")
	err = startInternal()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	fmt.Fprintln(os.Stdout, "openedge restarted")
}

func init() {
	restartCmd.Flags().StringVarP(&workDir, "workdir", "w", "", "work directory of openedge")
	restartCmd.Flags().StringVarP(&confFile, "config", "c", "", "config path of openedge")
	rootCmd.AddCommand(restartCmd)
}
