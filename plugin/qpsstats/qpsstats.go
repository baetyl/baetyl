// Package qpsstats qps监控实现
package qpsstats

import (
	goctx "context"
	"encoding/json"
	"fmt"
	gohttp "net/http"
	"strings"

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

var (
	lastQPSStatus = map[string]interface{}{}
)

func init() {
	v2plugin.RegisterFactory("qpsstats", New)
}

type qpsStats struct {
	cfg Config
	cli *client
	log *log.Logger
}

func (nv *qpsStats) Close() error {
	return nil
}

func New() (v2plugin.Plugin, error) {
	var cfg Config
	if err := utils.LoadYAML(plugin.ConfFile, &cfg); err != nil {
		return nil, errors.Trace(err)
	}
	cli, err := newClient(cfg.QPSStats.Kube)
	if err != nil {
		return nil, err
	}
	ns := &qpsStats{
		cfg: cfg,
		cli: cli,
		log: log.With(log.Any("plugin", "qps stats")),
	}
	return ns, nil
}

func (nv *qpsStats) CollectStats(_ string) (map[string]interface{}, error) {
	var res map[string]interface{}
	ns := context.EdgeNamespace()
	selector := "baetyl-webhook=true"

	pods, err := nv.cli.core.Pods(ns).List(goctx.TODO(), metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, errors.Trace(err)
	}
	if pods == nil || len(pods.Items) == 0 {
		nv.log.Info("QPS stats: no pod with label baetyl-webhook=true found in cluster")
		return nil, nil
	}
	for _, pod := range pods.Items {
		var collectPort int32
		podIP := pod.Status.PodIP
		for _, container := range pod.Spec.Containers {
			if container.Name == "sidecar-nginx" {
				collectPort = container.Ports[0].ContainerPort
			}
		}
		if collectPort == 0 {
			nv.log.Warn("labeled pod has no sidecar named nginx-sidecar or sidecar has no port config")
			continue
		}
		url := fmt.Sprintf("%s://%s:%d/%s", "http", podIP, collectPort, nv.cfg.QPSStats.CollectURL)
		log.L().Info("collect nginx sidecar status", log.Any("address", url))

		r, err := gohttp.DefaultClient.Get(url)
		if err != nil {
			return nil, errors.Trace(err)
		}
		data, err := http.HandleResponse(r)
		if err != nil {
			return nil, errors.Trace(err)
		}
		nginxStatus := NginxStatus{}
		if err = json.Unmarshal(data, &nginxStatus); err != nil {
			return nil, errors.Trace(err)
		}
		qpsStatus := generateQPSStats(nginxStatus)

		delta := generateDeltaQPSStats(nginxStatus.HostName, qpsStatus, lastQPSStatus)

		status := map[string]interface{}{
			nginxStatus.HostName: delta,
		}

		if err = mergo.Merge(&res, status); err != nil {
			return nil, errors.Trace(err)
		}
		lastQPSStatus[nginxStatus.HostName] = qpsStatus
	}
	return res, nil
}

func generateQPSStats(nginx NginxStatus) map[string]ServerStats {
	qps := make(map[string]ServerStats)
	for _, value := range nginx.UpstreamZones {
		ss := ServerStats{}
		serverName := strings.Split(value[0].Server, ":")
		ss.ServerName = serverName[len(serverName)-1]
		ss.RequestCnt = value[0].RequestCnt
		ss.RequestTotal = value[0].RequestCnt
		ss.RequestCntSuccess = value[0].Responses.Resp1xx + value[0].Responses.Resp2xx + value[0].Responses.Resp3xx
		ss.RequestCntFail = value[0].Responses.Resp4xx + value[0].Responses.Resp5xx

		qps[value[0].Server] = ss
	}
	return qps
}

func generateDeltaQPSStats(hostName string, qpsStats map[string]ServerStats, lastQPSStats map[string]interface{}) map[string]ServerStats {
	// first calculate
	if lastQPSStats[hostName] == nil {
		return qpsStats
	}
	last := lastQPSStats[hostName].(map[string]ServerStats)
	res := make(map[string]ServerStats)

	for key, value := range qpsStats {
		if _, ok := last[key]; ok && value.ServerName == last[key].ServerName && value.RequestCnt >= last[key].RequestCnt {
			deltaCnt := value.RequestCnt - last[key].RequestCnt
			deltaCntSuccess := value.RequestCntSuccess - last[key].RequestCntSuccess
			deltaCntFail := value.RequestCntFail - last[key].RequestCntFail
			res[key] = ServerStats{
				ServerName:        value.ServerName,
				RequestCnt:        deltaCnt,
				RequestTotal:      value.RequestTotal,
				RequestCntSuccess: deltaCntSuccess,
				RequestCntFail:    deltaCntFail,
			}
		} else {
			res[key] = ServerStats{
				ServerName:        value.ServerName,
				RequestCnt:        value.RequestCnt,
				RequestTotal:      value.RequestTotal,
				RequestCntSuccess: value.RequestCntSuccess,
				RequestCntFail:    value.RequestCntFail,
			}
		}
	}
	return res
}
