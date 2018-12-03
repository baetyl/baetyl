# OpenEdge

OpenEdge是开放的边缘计算平台，可将云计算能力拓展至用户现场，提供临时离线、低延时的计算服务，包括消息路由、函数计算、AI推断等。OpenEdge和[云端管理套件](https://cloud.baidu.com/product/bie.html)配合使用，可达到云端管理和应用下发，边缘设备上运行应用的效果，满足各种边缘计算场景。

**功能列表**：

> + 支持应用模块的管理，包括启停、重启、监听、守护和升级
> + 支持两种运行模式：Native进程模式和Docker容器模式
> + Docker容器模式支持资源隔离和资源限制
> + 支持云端管理套件，可以进行应用下发，设备信息上报等
> + 官方提供Hub模块，支持MQTT 3.1.1，支持QoS等级0和1，支持证书认证等
> + 官方提供函数计算模块，支持函数实例伸缩，支持SQL、Python2.7、AI推断等Runtime以及自定义Runtime
> + 官方提供远程服务通讯模块，支持MQTT协议
> + 官方提供视频流接入模块，支持RTMP
> + 提供模块SDK(Golang)，可用于开发自定义模块

**设计文档**：

> + [OpenEdge设计](./doc/openedge.md)
> + [所有配置解读](./doc/config.md)
> + [开发自定义模块](./doc/dev_module.md)

## 快速开始

> + [Linux环境](./doc/build_linux.md)

## 测试

    go test --race ./...

## 如何贡献

[TODO]

## 讨论

[TODO]
