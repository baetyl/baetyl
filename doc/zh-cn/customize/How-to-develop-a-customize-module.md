# 如何开发一个自定义功能模块

- [目录约定](#目录约定)
- [启动约定](#启动约定)
- [SDK](#sdk)

在开发自定义模块前请阅读[从源码编译 Baetyl](../setup/Build-from-Source.md)，了解 Baetyl 的编译环境要求。

自定义模块不限定开发语言，可运行即可，基本没有限制，甚至可以直接使用 `hub.docker.com` 上已有的镜像，比如 `eclipse-mosquitto`。但是了解下面介绍的约定，有利于更好、更快地开发自定义模块。

## 目录约定

目前，进程模式和容器模式一样，会为每个服务开辟独立的工作空间，虽然达不到隔离的效果，但是可以保证用户使用体验的一致。进程模式会在 `var/run/baetyl/services` 目录下为每个服务创建一个以服务名称命名的目录，服务程序启动时会指定该目录为工作目录，服务绑定的存储卷（volume）会映射（软链）到工作目录下。这里我们沿用容器模式的叫法，把该目录下的空间也称作容器，那么容器中的目录有如下推荐的使用方式：

- 容器中默认工作目录：`/`
- 容器中默认配置文件：`/etc/baetyl/service.yml`
- 容器中默认持久化路径：`/var/db/baetyl`
- 容器中默认日志路径：`/var/log/baetyl`

**注意**：如果数据需要持久化在设备（宿主机）上，比如数据库和日志，必须通过存储卷将容器中的目录映射到宿主机目录上，否者服务停止后数据会丢失。

## 启动约定

模块启动的方式没有过多要求，推荐从默认文件中加载YMAL格式的配置，然后运行模块的业务逻辑，最后监听 `SIGTERM` 信号来优雅退出。一个简单的 `Golang` 模块实现可参考 MQTT 远程通讯模块（`baetyl-remote-mqtt`）。

## SDK

如果模块使用 `Golang` 开发，可使用 Baetyl 提供的 SDK，位于该项目的 sdk 目录中，由 `Context` 提供功能接口。目前，提供的 SDK 能力还比较有限，后续会逐渐加强。

`Context` 接口列表如下：

```golang
// 返回服务的系统配置，比如 hub 和 logger
Config() *ServiceConfig
// 加载服务的自定义配置
LoadConfig(interface{}) error
// 通过系统配置创建一个连接 Hub 的 Client，可以指定 Client ID 和订阅的主题信息
NewHubClient(string, []mqtt.TopicInfo) (*mqtt.Dispatcher, error)
// 返回日志接口
Log() logger.Logger
// 等待退出，接收 SIGTERM 和 SIGINT 信号
Wait()
// 返回等待退出的 Channel
WaitChan() <-chan os.Signal

// 主程序 RESTful API

// 更新系统服务
UpdateSystem(string, string, string, bool) error
// 查看系统状态
InspectSystem() (*Inspect, error)
// 获取一个宿主机的空闲端口
GetAvailablePort() (string, error)
// 报告本实例的状态信息
ReportInstance(stats map[string]interface{}) error
// 启动某个服务的某个实例
StartInstance(serviceName, instanceName string, dynamicConfig map[string]string) error
// 停止某个服务的某个实例
StopInstance(serviceName, instanceName string) error
```

下面以简单定时器模块实现为例，介绍SDK的用法。

```golang
package main

import (
	"encoding/json"
	"time"

	"github.com/baetyl/baetyl/protocol/mqtt"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
)

// 自定义模块的自定义配置，
type config struct {
	Timer struct {
		Interval time.Duration `yaml:"interval" json:"interval" default:"1m"`
	} `yaml:"timer" json:"timer"`
	Publish mqtt.TopicInfo `yaml:"publish" json:"publish" default:"{\"topic\":\"timer\"}"`
}

func main() {
	// 模块在 Baetyl 的 Context 中启动，SDK 的功能均由 Context 提供
	baetyl.Run(func(ctx baetyl.Context) error {
		var cfg config
		// 加载自定义配置
		err := ctx.LoadConfig(&cfg)
		if err != nil {
			return err
		}
		// 创建连接Hub的客户端
		cli, err := ctx.NewHubClient("", nil)
		if err != nil {
			return err
		}
		// 启动客户端，支持自动重连
		cli.Start(nil)
		// 创建定时器
		ticker := time.NewTicker(cfg.Timer.Interval)
		defer ticker.Stop()
		for {
			select {
			case t := <-ticker.C:
				msg := map[string]int64{"time": t.Unix()}
				pld, _ := json.Marshal(msg)
				// 定时发送消息到 Hub
				err := cli.Publish(cfg.Publish, pld)
				if err != nil {
					// 打印错误日志
					ctx.Log().Errorf(err.Error())
				}
			case <-ctx.WaitChan():
				// 等待退出信号，SIGTERM 或者 SIGINT
				return nil
			}
		}
	})
}
```

`baetyl-timer` 的配置中，`hub` 属于系统配置，`timer` 和 `publish` 是该模块的自定义配置。

```yaml
hub:
  address: tcp://localhub:1883
  username: test
  password: hahaha
  clientid: timer1
timer:
  interval: 1s
publish:
  topic: timer1
```
