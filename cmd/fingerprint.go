package cmd

import (
	"fmt"

	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/spf13/cobra"
)

type GetFingerprintFunc func()

const (
	HookGetFingerprint = "getFingerprint"
)

func init() {
	rootCmd.AddCommand(fingerprint)
}

var fingerprint = &cobra.Command{
	Use:   "fingerprint",
	Short: "Obtain machine fingerprints.",
	Long:  "Observer machine fingerprints activation program",
	Run: func(_ *cobra.Command, _ []string) {
		Hooks[HookGetFingerprint].(GetFingerprintFunc)()
	},
}

func GetFingerprint() {
	info, err := utils.GetFingerprint("")
	if err != nil {
		log.L().Error("get fingerprint err: " + err.Error())
		return
	}
	fmt.Println("Copy The Fingerprint:")
	fmt.Println(info)
}
