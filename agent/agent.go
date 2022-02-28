package agent

import "github.com/qiangxue/fasthttp-routing"

//go:generate mockgen -destination=../mock/agent.go -package=mock -source=agent.go AgentClient

type AgentClient interface {
	SendRequest(ctx *routing.Context) (interface{}, error)
	GetOrSetAgentFlag(action string) (bool, error)
}
