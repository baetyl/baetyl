package cmd

import (
	"os"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/spf13/cobra"

	"github.com/baetyl/baetyl/v2/ami"
	"github.com/baetyl/baetyl/v2/config"
)

var (
	ns string
)

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringVarP(&ns, "namespace", "n", "baetyl-edge-system", "The namespace of applications which will be deleted.")
	deleteCmd.Flags().StringVarP(&mode, "mode", "m", "native", "The running mode of applications, supports 'kube' and 'native'.")
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete baetyl applications.",
	Long:  "Delete baetyl applications by namespace.",
	Run: func(_ *cobra.Command, _ []string) {
		delete()
	},
}

func delete() {
	var err error
	var l *log.Logger
	l, err = log.Init(log.Config{Level: "debug", Encoding: "console"})
	if err != nil {
		log.L().Error("failed to init logger", log.Error(err))
		return
	}
	defer func() {
		if err != nil {
			l.Error(err.Error())
		}
	}()

	// prepare env
	if _, ok := modes[mode]; !ok {
		err = errors.New("The parameter 'mode' is invalid")
		return
	}
	err = os.Setenv(context.KeyRunMode, mode)
	if err != nil {
		return
	}

	// stats app, then delete them
	amiConfig := config.AmiConfig{}
	err = utils.SetDefaults(&amiConfig)
	if err != nil {
		return
	}

	// TODO: create client like kubectl without confpath
	amiConfig.Kube.OutCluster = true
	amiConfig.Kube.ConfPath = ".kube/config"

	am, err := ami.NewAMI(mode, amiConfig)
	if err != nil {
		return
	}

	var allstats []specv1.AppStats
	allstats, err = am.StatsApps(ns)
	log.L().Info("stats apps", log.Any("all", allstats), log.Error(err))
	for _, appstats := range allstats {
		err = am.DeleteApp(ns, appstats.Name)
		if err != nil {
			log.L().Error("failed to delete app", log.Any("name", appstats.Name), log.Any("namespace", ns), log.Error(err))
		} else {
			log.L().Info("delete app", log.Any("name", appstats.Name), log.Any("namespace", ns))
		}
	}
}
