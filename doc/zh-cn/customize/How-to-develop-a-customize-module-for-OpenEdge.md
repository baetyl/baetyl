# 自定义模块

- [目录约定](#目录约定)
- [启动约定](#启动约定)
- [SDK](#sdk)

在开发和集成自定义模块前请阅读开发编译指南，了解 OpenEdge 的编译环境要求。

自定义模块不限定开发语言，可运行即可，基本没有限制，甚至可以直接使用hub.docker.com上已有的镜像，比如eclipse-mosquitto。但是了解下面介绍的约定，有利于更好更快的开发自定义模块。

## 目录约定

目前进程模式和容器模式一样，会为每个服务开辟独立的工作空间，虽然达不到隔离的效果，但是可以保证使用体验的一致。进程模式会在var/run/openedge/services目录下为每个服务创建一个以服务命名的目录，服务程序启动会指定该目录为工作目录，服务绑定的存储卷会映射（软链）到工作目录下。这里我们沿用容器模式的叫法，把该目录下的空间也叫容器，那么容器中的目录有如下推荐的使用方式：

容器中默认工作目录是：/
容器中默认配置文件是：/etc/openedge/service.yml
容器中默认持久化路径是：/var/db/openedge
容器中默认日志路径是：/var/log/openedge

**注意**：如果数据需要持久化在设备（宿主机）上，比如数据库和日志，必须通过存储卷将容器中的目录映射到宿主机目录上，否者服务停止后数据会丢失。

## 启动约定

模块启动的方式没有过多要求，推荐从默认文件中加载 YMAL 格式的配置，然后运行模块的业务逻辑，最后监听 `SIGTERM` 信号来优雅退出。一个简单的 `Golang` 模块实现可参考MQTT远程通讯模块（openedge-remote-mqtt）。

## SDK

如果模块使用 `Golang` 开发，可使用 OpenEdge 提供的 SDK，位于该项目的sdk目录中。下面以简单定时器模块实现为例，介绍SDK的用法，当然目前提供的SDK能力还比较有限，后续会逐渐加强。

openedge-timer的golang实现。

```golang
package main

import (
	"encoding/json"
	"time"

	"github.com/baidu/openedge/protocol/mqtt"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
)

// 自定义模块的自定义配置，
type config struct {
	Timer struct {
		Interval time.Duration `yaml:"interval" json:"interval" default:"1m"`
	} `yaml:"timer" json:"timer"`
	Publish mqtt.TopicInfo `yaml:"publish" json:"publish" default:"{\"topic\":\"timer\"}"`
}

func main() {
	// 模块在OpenEdge的Context中启动，主要SDK的只要功能都有Context提供
	openedge.Run(func(ctx openedge.Context) error {
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
				// 定时发送消息到Hub
				err := cli.Publish(cfg.Publish, pld)
				if err != nil {
					// 打印错误日志
					ctx.Log().Errorf(err.Error())
				}
			case <-ctx.WaitChan():
				// 等待退出信号，SIGTERM或者SIGINT
				return nil
			}
		}
	})
}
```

openedge-timer的配置，hub属于系统配置，timer和pubish是该模块的自定义配置。

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
