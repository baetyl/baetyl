# Pre-release 0.1.2(2019-04-04)

## Features

- 从主程序分离 Agent 模块，定时上报（核心）设备状态信息
- 针对资源配置引入数据卷（volume）概念，灵活配置，支持第三方已有的镜像，比如 `hub.docker.com` 中的 mosquitto
- 支持命令行启动（后台以服务方式运行）、停止 OpenEdge 服务
- 统一两种模式（Docker 容器模式 和 Native 进程模式）配置，例如为本机 Native 进程模式中的每个服务创建单独的工作目录
- 引入服务概念，代替模块，用于支持启动多实例
- 针对 Docker 容器模式支持设备映射

## Bug fixes

- 支持 `openedge.sock` 清理逻辑
- 升级 Hub 模块的连接鉴权、认证逻辑，如 TCP 连接采用明文存储密码
- 升级函数计算模块，支持重试逻辑，去除保序逻辑

## Others(include release engineering)

- 丰富的官方测试案例支持，如针对 Hub 模块，提供基于 mosquitto 的配置
- 全量文档支持英文