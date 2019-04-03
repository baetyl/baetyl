# OpenEdge 构成

一个完整的 OpenEdge 系统由主程序、服务、存储卷和使用的系统资源组成，需要注意的是同一个服务下的实例共享所有存储卷，所以如果出现独占的资源，比如监听同一个端口，使用同一个Client ID 连接 Hub 等，只能启动一个实例。各相关组成部分的主要功能如下：

- **主程序**：作为 OpenEdge 系统的核心，主程序负责管理所有存储卷和服务，内置运行引擎系统，对外提供 RESTful API 和命令行。
	- **引擎系统**：负责服务的存储卷映射，实例启停、监听和守护等。目前支持了 Docker 容器模式和 Native 进程模式，后续还会支持 k3s 容器模式；
	- **RESTful API**：OpenEdge 主程序会暴露一组 RESTful API，在 Linux 系统下默认采用 Unix Domain Socket，固定地址为 `/var/openedge.sock`；其他环境采用 TCP，默认地址为 `tcp://127.0.0.1:50050`；
	- **命令行**：OpenEdge 支持以命令行发布，启动，后台服务运行，目前可支持 `start`、`version`、`help`及`stop` 几个命令。
- **服务**：指一组接受 OpenEdge 控制的运行程序集合，用于提供某些具体的功能，比如消息路由服务、函数计算服务、微服务等。
	- **Agent 模块**：提供 BIE 云代理服务，负责边缘核心设备状态上报和应用下发；
	- **Hub 模块**：提供基于 [MQTT 协议](http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.html) 的订阅和发布功能，支持TCP、SSL（TCP+SSL）、WS（Websocket）及 WSS（Websocket+SSL）四种接入方式、消息路由转发等功能；
	- **函数计算模块**：提供基于 MQTT 消息机制，弹性、高可用、扩展性好、响应快的的计算能力，并且兼容[百度云-函数计算 CFC](https://cloud.baidu.com/product/cfc.html)。需要注意的是函数计算不保证消息顺序，除非只启动一个函数实例。特别地，基于函数计算模块还提供 Python2.7 运行时模块，用于加载 Python 脚本的 GRPC 微服务，可以托管给函数计算模块成为函数实例提供方；
	- **远程通讯模块**：目前支持 MQTT 协议，其实质是两个 MQTT Server 的桥接（Bridge）模块，用于订阅一个 Server 的消息并转发给另一个 Server；目前支持配置多路消息转发，可配置多个 Remote 和 Hub 同时进行消息同步。
- **存储卷**：指被服务使用的目录，可以是只读的目录，比如放置配置、证书、脚本等资源的目录，也可以是可写的目录，比如日志、数据等持久化目录。