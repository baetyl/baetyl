package cmd

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/spf13/cobra"

	"github.com/baetyl/baetyl/v2/ami"
	"github.com/baetyl/baetyl/v2/config"
	"github.com/baetyl/baetyl/v2/sync"
)

var (
	file       string
	mode       string
	skipVerify bool
	modes      = map[string]struct{}{
		context.RunModeKube:   {},
		context.RunModeNative: {},
	}
)

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().StringVarP(&file, "filename", "f", "", "The application mode file to apply, only support json format now.")
	applyCmd.Flags().StringVarP(&mode, "mode", "m", "native", "The running mode of applications, supports 'kube' and 'native'.")
	applyCmd.Flags().BoolVar(&skipVerify, "skip-verify", false, "Indicates whether to skip certificate verify.")
	applyCmd.MarkFlagRequired("filename")
}

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply baetyl applications.",
	Long:  "Apply baetyl applications, can run applications in kube mode or native mode.",
	Run: func(_ *cobra.Command, _ []string) {
		apply()
	},
}

func apply() {
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
		err = errors.New("The mode is invalid")
		return
	}
	err = os.Setenv(context.KeyRunMode, mode)
	if err != nil {
		return
	}
	hostPathLib, err := context.HostPathLib()
	if err != nil {
		return
	}
	var data []byte
	// download data
	ops := http.NewClientOptions()
	ops.Timeout = 10 * time.Minute
	if skipVerify {
		ops.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}
	cli := http.NewClient(ops)

	if isValidUrl(file) {
		data, err = cli.GetJSON(file)
		if err != nil {
			return
		}
	} else {
		data, err = ioutil.ReadFile(file)
		if err != nil {
			return
		}
	}
	data, err = utils.ParseEnv(data)
	if err != nil {
		return
	}

	var all []v1.ResourceValue
	err = json.Unmarshal(data, &all)
	if err != nil {
		return
	}

	apps := map[string]v1.Application{}
	configs := map[string]v1.Configuration{}
	secrets := map[string]v1.Secret{}

	for _, r := range all {
		if app := r.App(); app != nil {
			apps[app.Name] = *app
		} else if conf := r.Config(); conf != nil {
			configs[conf.Name] = *conf
		} else if secret := r.Secret(); secret != nil {
			secrets[secret.Name] = *secret
		}
	}

	// download config objects if exist
	objectHostPath := filepath.Join(hostPathLib, "object")
	for _, cfg := range configs {
		sync.FilterConfig(&cfg)
		err = sync.DownloadConfig(cli, objectHostPath, &cfg)
		if err != nil {
			return
		}
	}

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
	for _, app := range apps {
		// prepare host path
		for _, v := range app.Volumes {
			if v.HostPath != nil {
				err = os.MkdirAll(v.HostPath.Path, 0766)
				if err != nil {
					return
				}
			}
		}

		// prepare app
		hostHostPath := filepath.Join(hostPathLib, "host")
		err = sync.PrepareApp(hostHostPath, objectHostPath, &app, configs)
		if err != nil {
			return
		}

		// apply app
		err = am.ApplyApp(app.Namespace, app, configs, secrets)
		if err != nil {
			return
		}
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	signal.Ignore(syscall.SIGPIPE)
	t := time.NewTicker(time.Second * 6)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			allstats, err := am.StatsApps(context.EdgeSystemNamespace())
			log.L().Info("stats system apps", log.Any("all", allstats), log.Error(err))
			if allstats == nil {
				continue
			}
			success := true
		loop:
			for _, appstats := range allstats {
				for _, insstats := range appstats.InstanceStats {
					if insstats.Status != v1.Running {
						success = false
						break loop
					}
				}
			}
			if success {
				log.L().Info("baetyl apply application(s) successfully")
				return
			}
		case <-sig:
			return
		}
	}
}

func isValidUrl(urls string) bool {
	_, err := url.ParseRequestURI(urls)
	if err != nil {
		return false
	}

	u, err := url.Parse(urls)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}
