package engine

import (
	"fmt"
	"io"
	"strings"

	openedge "github.com/baidu/openedge/sdk/openedge-go"
)

// all status
const (
	KeyID         = "id"
	KeyName       = "name"
	KeyStatus     = "status"
	KeyCreateTime = "create_time"
	KeyStartTime  = "start_time"
	KeyFinishTime = "finish_time"

	// Created    = "created"    // 已创建
	Running = "running" // 运行中
	// Paused     = "paused"     // 已暂停
	Restarting = "restarting" // 重启中
	// Removing   = "removing"   // 退出中
	// Exited     = "exited"     // 已退出
	Dead = "dead" // 未启动（默认值）
	// Offline    = "offline"    // 离线（同核心的状态）
)

// Instance interfaces of instance
type Instance interface {
	ID() string
	Name() string
	Service() Service
	Wait(w chan<- error)
	Dying() <-chan struct{}
	Restart() error
	Stop()
	io.Closer
}

// GenerateInstanceEnv generates new env of the instance
func GenerateInstanceEnv(name string, static []string, dynamic map[string]string) []string {
	env := []string{}
	dyn := dynamic != nil
	for _, v := range static {
		// remove auth token info for dynamic instances
		if dyn && strings.HasPrefix(v, openedge.EnvServiceTokenKey) {
			continue
		}
		env = append(env, v)
	}
	env = append(env, fmt.Sprintf("%s=%s", openedge.EnvServiceInstanceNameKey, name))
	if dyn {
		for k, v := range dynamic {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
	}
	return env
}
