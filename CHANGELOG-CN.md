# Pre-release 0.1.2(2019-04-04)

## Features

- 从主程序分离 Agent 模块，定时上报（核心）设备状态信息
- 针对资源配置引入存储卷（volume）概念，灵活配置，支持第三方已有的镜像，比如 `hub.docker.com` 中的 mosquitto
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

# Pre-release 0.1.1(2018-12-28)

## 功能

- 优化 MQTT 通讯模块，支持配置多路消息转发，可配置多个 Remote 和 Hub 同时进行消息同步
- 增加合法订阅主题配置，用于校验 MQTT 客户端订阅结果

## Bug 修复

- Docker 容器模式下模块目录隔离
- 移除网络范围过滤器，以支持旧版本 Docker

## 其他

- 更丰富的构建、测试、发布脚本和文档
- 引入 vendor 依赖包，有效解决所有编译依赖问题
- 重构代码并格式化所有消息
- 使用 Makefile 代替 Shell 脚本编译 OpenEdge
- 更新 gomqtt
- 增加 travis 持续集成服务

# Pre-release 0.1.0(2018-12-05)

百度边缘计算产品 OpenEdge 正式宣布开源。

## 功能

- 完成模块化改造、支持模块管理
- 支持两种运行模式：Docker 容器模式和 Native 进程模式
- Docker 容器模式支持资源隔离和限制（比如 CPU、内存等）
- 提供诸如本地 Hub、本地函数计算（包含 Python2.7 运行时）、MQTT 远程通讯模块等

## Bug 修复

- N/A