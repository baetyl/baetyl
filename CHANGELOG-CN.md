# Pre-release 0.1.4(2019-07-05)

## 功能

- [#251](https://github.com/baetyl/baetyl/issues/251) 增加 Nodejs85 函数计算运行时
- [#260](https://github.com/baetyl/baetyl/issues/260) 主程序收集 IP 和 MAC 信息
- [#263](https://github.com/baetyl/baetyl/issues/263) 优化应用 OTA，不再重启配置未变化的服务
- [#264](https://github.com/baetyl/baetyl/issues/264) 优化存储卷清理逻辑，将其从 主程序迁移至 Agent 模块，会清除所有不在应用的存储卷列表中的目录
- [#266](https://github.com/baetyl/baetyl/issues/266) 采集服务实例的 CPU 与内存资源使用信息

## Bug 修复

- [#246](https://github.com/baetyl/baetyl/issues/246) Agent 模块状态上报周期从 1 分钟变更为 20 秒

## 其他

- [#269](https://github.com/baetyl/baetyl/issues/269) [#273](https://github.com/baetyl/baetyl/issues/273) [#280](https://github.com/baetyl/baetyl/issues/280) 更新 Makefile 文件，支持选择性部署

# Pre-release 0.1.3(2019-05-10)

## 功能

- [#199](https://github.com/baetyl/baetyl/issues/199) 支持上报服务实例的自定义状态信息，同时采集更多系统状态信息，具体参考 [Baetyl 系统信息采集](./doc/zh-cn/overview/Design.md#system-inspect)
- [#209](https://github.com/baetyl/baetyl/issues/209) 当 baetyl 启动时，会清理旧实例（残留实例一般由于 baetyl 异常退出造成）
- [#211](https://github.com/baetyl/baetyl/issues/211) 增加了 Python3.6 版本的函数运行时，使用 Ubuntu16.04 作为基础镜像
- [#222](https://github.com/baetyl/baetyl/issues/222) Docker 容器模式支持 runtime 和 args 配置
- 添加了一个简单的定时器模块 baetyl-timer

## Bug 修复

- [#201](https://github.com/baetyl/baetyl/issues/201) 函数实例池销毁函数实例时，确保停止函数实例
- [#208](https://github.com/baetyl/baetyl/issues/208) baetyl stop 命令等待 baetyl 停止运行后退出，确保清理pid文件
- [#234](https://github.com/baetyl/baetyl/issues/234) hub 模块发布消息等待 ack 超时，快速重发消息
- 解决当 atomic.addUint64() 的参数未按照 64 位对齐导致退出异常的问题。参考：https://github.com/golang/go/issues/23345

## 其他

- [#230](https://github.com/baetyl/baetyl/issues/230) 发布 Baetyl 2019 Roadmap
- [#228](https://github.com/baetyl/baetyl/issues/228) 发布 Baetyl 社区参与者公约

# Pre-release 0.1.2(2019-04-04)

## 功能

- [#20](https://github.com/baetyl/baetyl/issues/20) 从主程序分离 Agent 模块，定时上报（核心）设备状态信息
- [#120](https://github.com/baetyl/baetyl/issues/120) 针对资源配置引入存储卷（volume）概念，灵活配置，支持第三方已有的镜像，比如 `hub.docker.com` 中的 mosquitto
- [#122](https://github.com/baetyl/baetyl/issues/122) 支持命令行启动（后台以服务方式运行）、停止 Baetyl 服务
- 统一两种模式（Docker 容器模式 和 Native 进程模式）配置，例如为本机 Native 进程模式中的每个服务创建单独的工作目录
- [#123](https://github.com/baetyl/baetyl/issues/123) 引入服务概念，代替模块，用于支持启动多实例
- [#142](https://github.com/baetyl/baetyl/issues/142) 针对 Docker 容器模式支持设备映射

## Bug 修复

- [#92](https://github.com/baetyl/baetyl/issues/92) [#81](https://github.com/baetyl/baetyl/issues/81) 支持 `baetyl.sock` 清理逻辑
- [#135](https://github.com/baetyl/baetyl/issues/135) [#88](https://github.com/baetyl/baetyl/issues/88) 升级 Hub 模块的连接鉴权、认证逻辑，如 TCP 连接采用明文存储密码
- [#127](https://github.com/baetyl/baetyl/issues/127) 升级函数计算模块，支持重试逻辑，去除保序逻辑

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
- 使用 Makefile 代替 Shell 脚本编译 Baetyl
- 更新 gomqtt
- 增加 travis 持续集成服务

# Pre-release 0.1.0(2018-12-05)

百度边缘计算产品 Baetyl 正式宣布开源。

## 功能

- 完成模块化改造、支持模块管理
- 支持两种运行模式：Docker 容器模式和 Native 进程模式
- Docker 容器模式支持资源隔离和限制（比如 CPU、内存等）
- 提供诸如本地 Hub、本地函数计算（包含 Python2.7 运行时）、MQTT 远程通讯模块等

## Bug 修复

- N/A