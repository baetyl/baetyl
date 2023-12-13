// Package agent 代理功能实现
package agent

import (
	"encoding/json"
	gosync "sync"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/qiangxue/fasthttp-routing"

	"github.com/baetyl/baetyl/v2/sync"
)

//go:generate mockgen -destination=../mock/agent.go -package=mock -source=agent.go AgentClient

type AgentClient interface {
	SendRequest(ctx *routing.Context) (interface{}, error)
	GetOrSetAgentFlag(action string) (bool, error)
}

var ErrAgentNotStart = errors.New("failed to request, agent is closed")

type agentClient struct {
	syn  sync.Sync
	res  *v1.STSResponse
	flag bool
	gosync.RWMutex
}

func NewAgentClient(syn sync.Sync) (AgentClient, error) {
	return &agentClient{
		syn:  syn,
		flag: true,
	}, nil
}

func (a *agentClient) SendRequest(ctx *routing.Context) (interface{}, error) {
	a.RLock()
	if !a.flag {
		a.RUnlock()
		http.RespondMsg(ctx, 400, "agent is closed", ErrAgentNotStart.Error())
		return nil, errors.Trace(ErrAgentNotStart)
	}
	if a.res != nil && a.res.Expiration.UTC().Sub(time.Now().UTC()) >= time.Hour {
		a.RUnlock()
		return a.res, nil
	}
	a.RUnlock()
	var stsReq v1.STSRequest
	if err := json.Unmarshal(ctx.Request.Body(), &stsReq); err != nil || stsReq.STSType != v1.MessageSTSTypeMinio {
		http.RespondMsg(ctx, 500, "invalid input", string(ctx.Request.Body()))
		return nil, errors.Trace(ErrAgentNotStart)
	}
	stsReq.ExpiredTime = 12 * time.Hour
	res, err := a.syn.Request(&v1.Message{
		Kind: v1.MessageSTS,
		Metadata: map[string]string{
			string(v1.MessageCMD): v1.MessageSTSTypeMinio,
		},
		Content: v1.LazyValue{Value: stsReq},
	})
	if err != nil {
		http.RespondMsg(ctx, 400, "failed to get sts from cloud", err.Error())
		return nil, errors.Trace(err)
	}
	var stsToken v1.STSResponse
	err = res.Content.Unmarshal(&stsToken)
	if err != nil {
		http.RespondMsg(ctx, 500, "failed to allocate sts", err.Error())
		return nil, errors.Trace(err)
	}
	log.L().Debug("Request sts success", log.Any("body", string(ctx.Request.Body())))
	a.Lock()
	a.res = &stsToken
	a.Unlock()
	return stsToken, nil
}

func (a *agentClient) GetOrSetAgentFlag(action string) (bool, error) {
	a.Lock()
	defer a.Unlock()
	switch action {
	case "start":
		a.flag = true
	case "close":
		a.flag = false
	}
	return a.flag, nil
}
