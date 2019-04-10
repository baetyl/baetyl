package engine

import (
	"fmt"
	"io"
	"strings"
	"sync"

	openedge "github.com/baidu/openedge/sdk/openedge-go"
)

// all status
const (
	Created    = "created"    // 已创建
	Running    = "running"    // 运行中
	Paused     = "paused"     // 已暂停
	Restarting = "restarting" // 重启中
	Removing   = "removing"   // 退出中
	Exited     = "exited"     // 已退出
	Dead       = "dead"       // 未启动（默认值）
	Offline    = "offline"    // 离线（同核心的状态）
)

// Instance interfaces of instance
type Instance interface {
	ID() string
	Name() string
	Supervisee
	io.Closer
}

// InstanceStats instance stats
type InstanceStats struct {
	stats map[string]interface{}
	mutex sync.RWMutex
}

// ID gets the id of the instance
func (i *InstanceStats) ID() string {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	return i.stats["id"].(string)
}

// Name gets the name of the instance
func (i *InstanceStats) Name() string {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	return i.stats["name"].(string)
}

// Status gets the status of the instance
func (i *InstanceStats) Status() string {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	return i.stats["status"].(string)
}

// Stat gets a stat of the instance
func (i *InstanceStats) Stat(k string) interface{} {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	return i.stats[k]
}

// Stats gets the stats of the instance
func (i *InstanceStats) Stats() map[string]interface{} {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	stats := map[string]interface{}{}
	for k, v := range i.stats {
		stats[k] = v
	}
	return stats
}

// SetStatus sets the status of the instance
func (i *InstanceStats) SetStatus(v interface{}) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	if i.stats == nil {
		i.stats = map[string]interface{}{}
	}
	i.stats["status"] = v
}

// SetError sets the error of the instance
func (i *InstanceStats) SetError(v interface{}) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	if i.stats == nil {
		i.stats = map[string]interface{}{}
	}
	i.stats["error"] = v
}

// SetStat sets a stat of the instance
func (i *InstanceStats) SetStat(k string, v interface{}) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	if i.stats == nil {
		i.stats = map[string]interface{}{}
	}
	i.stats[k] = v
}

// SetStats sets the stats of the instance
func (i *InstanceStats) SetStats(kvs map[string]interface{}) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	if i.stats == nil {
		i.stats = map[string]interface{}{}
	}
	for k, v := range kvs {
		i.stats[k] = v
	}
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
