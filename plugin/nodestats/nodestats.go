// Package nodestats 节点状态监控实现
package nodestats

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
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/baetyl/baetyl/v2/plugin"
	"github.com/imdario/mergo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	v2plugin.RegisterFactory("nodestats", New)
}

type nodeStats struct {
	cfg Config
	cli *client
	log *log.Logger
}

func (nv *nodeStats) Close() error {
	return nil
}

func New() (v2plugin.Plugin, error) {
	var cfg Config
	if err := utils.LoadYAML(plugin.ConfFile, &cfg); err != nil {
		return nil, errors.Trace(err)
	}
	cli, err := newClient(cfg.NodeStats.Kube)
	if err != nil {
		return nil, err
	}
	ns := &nodeStats{
		cfg: cfg,
		cli: cli,
		log: log.With(log.Any("plugin", "node stats")),
	}
	return ns, nil
}

func (nv *nodeStats) CollectStats(_ string) (map[string]interface{}, error) {
	var res map[string]interface{}
	ns := context.EdgeSystemNamespace()
	selector := "baetyl-service-name=baetyl-agent"

	pods, err := nv.cli.core.Pods(ns).List(goctx.TODO(), metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, errors.Trace(err)
	}
	if pods == nil || len(pods.Items) == 0 {
		nv.log.Info("disk&net stats: no pod with label baetyl-service-name=baetyl-agent found in cluster")
		return nil, nil
	}

	for _, pod := range pods.Items {
		podIP := pod.Status.PodIP
		url := fmt.Sprintf("%s://%s:%s%s", "http", podIP, nv.cfg.NodeStats.CollectPort, nv.cfg.NodeStats.CollectURL)
		log.L().Info("collect node stats pod", log.Any("address", url))

		r, err := gohttp.DefaultClient.Get(url)
		if err != nil {
			return nil, errors.Trace(err)
		}
		data, err := http.HandleResponse(r)
		if err != nil {
			return nil, errors.Trace(err)
		}
		var tmp map[string]interface{}
		if err = json.Unmarshal(data, &tmp); err != nil {
			return nil, errors.Trace(err)
		}
		if err = mergo.Merge(&res, tmp); err != nil {
			return nil, errors.Trace(err)
		}
	}
	return res, nil
}
