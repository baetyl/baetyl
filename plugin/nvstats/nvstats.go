// Package nvstats nv 状态监控实现
package nvstats

import (
	goctx "context"
	"encoding/json"
	"fmt"
	gohttp "net/http"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	v2plugin "github.com/baetyl/baetyl-go/v2/plugin"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/baetyl/baetyl/v2/plugin"
	"github.com/imdario/mergo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	BaetylServiceName = "baetyl-service-name"
)

func init() {
	v2plugin.RegisterFactory("nvstats", New)
	v2plugin.RegisterFactory("nativenvstats", NewNative)
}

type nvStats struct {
	cli  *client
	cfg  Config
	http *http.Client
	log  *log.Logger
}

func (nv *nvStats) Close() error {
	return nil
}

func New() (v2plugin.Plugin, error) {
	var cfg Config
	if err := utils.LoadYAML(plugin.ConfFile, &cfg); err != nil {
		return nil, errors.Trace(err)
	}
	cli, err := newClient(cfg.NvStats.KubeConfig)
	if err != nil {
		return nil, err
	}
	nv := &nvStats{
		cli: cli,
		cfg: cfg,
		log: log.With(log.Any("plugin", "gpu stats")),
	}
	return nv, nil
}

func NewNative() (v2plugin.Plugin, error) {
	var cfg Config
	if err := utils.LoadYAML(plugin.ConfFile, &cfg); err != nil {
		return nil, errors.Trace(err)
	}
	nv := &nvStats{
		cfg: cfg,
		log: log.With(log.Any("plugin", "native gpu stats")),
	}
	return nv, nil
}

// since each gpu node has a gpu metrics daemonset and a pod,
// all node gpu stats can be collected from each service running in the pod
func (nv *nvStats) CollectStats(mode string) (map[string]interface{}, error) {
	if mode == context.RunModeKube {
		return nv.collectKubeStats()
	} else if mode == context.RunModeNative {
		return nv.collectNativeStats()
	}
	return nil, nil
}

func (nv *nvStats) collectKubeStats() (map[string]interface{}, error) {
	var nodesStats map[string]interface{}
	ns := context.EdgeSystemNamespace()
	pods, err := nv.cli.core.Pods(ns).List(goctx.TODO(), metav1.ListOptions{LabelSelector: labels.FormatLabels(map[string]string{BaetylServiceName: v1.BaetylGPUMetrics})})
	if err != nil {
		return nil, errors.Trace(err)
	}
	if pods == nil || len(pods.Items) == 0 {
		nv.log.Info("gpu stats: no pod with label baetyl-service-name=baetyl-accelerator-metrics found in cluster")
		return nil, nil
	}
	for _, item := range pods.Items {
		url := fmt.Sprintf("http://%s:%d%s", item.Status.PodIP,
			nv.cfg.NvStats.CollectPort, nv.cfg.NvStats.CollectURL)
		log.L().Info("collect gpu stats pod", log.Any("address", url))

		resp, err := gohttp.DefaultClient.Get(url)
		if err != nil {
			return nil, errors.Trace(err)
		}
		data, err := http.HandleResponse(resp)
		if err != nil {
			return nil, errors.Trace(err)
		}
		var stats map[string]interface{}
		if err = json.Unmarshal(data, &stats); err != nil {
			return nil, errors.Trace(err)
		}
		if err = mergo.Merge(&nodesStats, stats); err != nil {
			return nil, errors.Trace(err)
		}
	}
	return nodesStats, nil
}

func (nv *nvStats) collectNativeStats() (map[string]interface{}, error) {
	url := fmt.Sprintf("http://%s:%d%s", "127.0.0.1",
		nv.cfg.NvStats.CollectPort, nv.cfg.NvStats.CollectURL)
	log.L().Info("collect gpu stats process", log.Any("address", url))

	resp, err := gohttp.DefaultClient.Get(url)
	if err != nil {
		return nil, errors.Trace(err)
	}
	data, err := http.HandleResponse(resp)
	if err != nil {
		return nil, errors.Trace(err)
	}
	var stats map[string]interface{}
	if err = json.Unmarshal(data, &stats); err != nil {
		return nil, errors.Trace(err)
	}
	return stats, nil
}
