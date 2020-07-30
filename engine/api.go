package engine

import (
	"encoding/json"
	"runtime"
	"strings"
	"time"

	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	routing "github.com/qiangxue/fasthttp-routing"
)

func (e *Engine) CollectReport(ctx *routing.Context) error {
	nodeInfo, err := e.ami.CollectNodeInfo()
	if err != nil {
		e.log.Warn("failed to collect node info", log.Error(err))
		http.RespondMsg(ctx, 500, "UnknownError", err.Error())
		return nil
	}
	nodeStats, err := e.ami.CollectNodeStats()
	if err != nil {
		e.log.Warn("failed to collect node stats", log.Error(err))
		http.RespondMsg(ctx, 500, "UnknownError", err.Error())
		return nil
	}
	appStats, err := e.ami.CollectAppStats(e.ns)
	if err != nil {
		e.log.Warn("failed to collect app stats", log.Error(err))
		http.RespondMsg(ctx, 500, "UnknownError", err.Error())
		return nil
	}
	sysappStats, err := e.ami.CollectAppStats(e.sysns)
	if err != nil {
		e.log.Warn("failed to collect system app stats", log.Error(err))
		http.RespondMsg(ctx, 500, "UnknownError", err.Error())
		return nil
	}
	apps := make([]specv1.AppInfo, 0)
	for _, info := range appStats {
		app := specv1.AppInfo{
			Name:    info.Name,
			Version: info.Version,
		}
		apps = append(apps, app)
	}
	sysapps := make([]specv1.AppInfo, 0)
	for _, info := range sysappStats {
		sysapp := specv1.AppInfo{
			Name:    info.Name,
			Version: info.Version,
		}
		sysapps = append(sysapps, sysapp)
	}
	r := specv1.Report{
		"time":      time.Now(),
		"node":      nodeInfo,
		"nodestats": nodeStats,
		"core": specv1.CoreInfo{
			GoVersion:   runtime.Version(),
			BinVersion:  utils.VERSION,
			GitRevision: utils.REVISION,
		},
	}
	r.SetAppInfos(false, apps)
	r.SetAppStats(false, appStats)
	r.SetAppInfos(true, sysapps)
	r.SetAppStats(true, sysappStats)
	data, err := json.Marshal(r)
	if err != nil {
		e.log.Error("failed to marshal report", log.Error(err))
		http.RespondMsg(ctx, 500, "UnknownError", err.Error())
		return nil
	}
	http.RespondStream(ctx, 200, strings.NewReader(string(data)), -1)
	return nil
}

func (e *Engine) GetServiceLog(ctx *routing.Context) error {
	service := ctx.Param("service")
	isSys := string(ctx.QueryArgs().Peek("system"))
	tailLines := string(ctx.QueryArgs().Peek("tailLines"))
	sinceSeconds := string(ctx.QueryArgs().Peek("sinceSeconds"))

	tail, since, err := e.validParam(tailLines, sinceSeconds)
	if err != nil {
		http.RespondMsg(ctx, 400, "RequestParamInvalid", err.Error())
		return nil
	}
	ns := e.ns
	if isSys == "true" {
		ns = e.sysns
	}
	reader, err := e.ami.FetchLog(ns, service, tail, since)
	if err != nil {
		http.RespondMsg(ctx, 500, "UnknownError", err.Error())
		return nil
	}
	http.RespondStream(ctx, 200, reader, -1)
	return nil
}
