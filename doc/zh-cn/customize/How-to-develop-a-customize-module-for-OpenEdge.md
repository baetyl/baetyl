# 自定义模块

- [自定义模块](#自定义模块)
  - [目录约定](#目录约定)
    - [Docker容器模式](#docker容器模式)
    - [Native进程模式](#native进程模式)
  - [配置约定](#配置约定)
  - [启动约定](#启动约定)
  - [模块SDK](#模块sdk)

在开发和集成自定义模块前请阅读开发编译指南，了解 OpenEdge 的编译环境要求。

自定义模块不限定开发语言，可运行即可。了解下面的这些约定，有利于更好更快的集成自定义模块。

## 目录约定

### Docker 容器模式

容器中的工作目录是：/

容器中的配置路径是：/etc/openedge/module.yml

容器中的配置资源读取目录是：/var/db/openedge/module/<模块名>

容器中的持久化数据输出目录是：/var/db/openedge/volume/<模块名>

容器中的持久化日志输出目录是：/var/log/openedge/<模块名>

具体的文件映射如下：

> - <openedge_host_work_dir>/var/db/openedge/module/<module_name>/module.yml:/etc/openedge/module.yml
> - <openedge_host_work_dir>/var/db/openedge/module/<module_name>:/var/db/openedge/module/<module_name>
> - <openedge_host_work_dir>/var/db/openedge/volume/<module_name>:/var/db/openedge/volume/<module_name>
> - <openedge_host_work_dir>/var/log/openedge/<module_name>:/var/log/openedge/<module_name>

**注意**：如果数据需要持久化在设备（宿主机）上，比如数据库和日志，必须保存在上述指定的持久化目录中，否者容器销毁数据会丢失。

### Native 进程模式

模块的工作目录和 OpenEdge 主程序的工作目录相同。

模块的配置路径是：<openedge_host_work_dir>/var/db/openedge/module/<模块名>/module.yml

模块的配置资源读取目录是：<openedge_host_work_dir>/var/db/openedge/module/<模块名>

模块的数据输出目录是：<openedge_host_work_dir>/var/db/openedge/volume/<模块名>

模块的日志输出目录是：<openedge_host_work_dir>/var/log/openedge/<模块名>

## 配置约定

模块支持从文件中加载 `YAML` 格式的配置，Docker 容器模式下读取`/etc/openedge/module.yml`，Native 进程模式下读取 `<openedge_host_work_dir>/var/db/openedge/module/<模块名>/module.yml`。

也支持从输入参数中获取配置，可以是 `JSON` 格式的字符串。比如：

    modules:
      - name: 'my_module'
        entry: 'my_module_docker_image'
        params:
          - '-c'
          - '{"name":"my_module","address":"127.0.0.1:1234",...}'

也支持从环境变量中获取配置。比如：

    modules:
      - name: 'my_module'
        entry: 'my_module_docker_image'
        env:
          name: my_module
          address: '127.0.0.1:1234'
          ...

## 启动约定

模块以进程方式启动，是独立的可执行程序。首先配置的约定方式加载配置，然后运行模块的业务逻辑，最后监听 `SIGTERM` 信号来优雅退出。一个简单的 `Golang` 模块实现可参考 [MQTT Remote 模块](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/openedge-remote-mqtt)。

## 模块 SDK

如果模块使用 `Golang` 开发，可使用 OpenEdge 中提供的部分 SDK，位于 [module](github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/module)。

[mqtt.Dispatcher](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/module/mqtt/dispatcher.go) 可用于向 MQTT Hub 订阅消息，支持自动重连。该 MQTT Dispatcher 不支持消息持久化，消息持久化应由 MQTT Hub 负责，MQTT Dispatcher 订阅的消息如果QoS 为1，则消息处理完后需要主动回复 ACK，否者 Hub 会重发消息。[MQTT Remote 模块参考](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/openedge-remote-mqtt/main.go)

[runtime.Server](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/module/function/runtime/server.go) 封装了 GRPC Server 和函数计算调用逻辑，方便开发者实现消息处理的函数计算运行时。[Python2.7 运行时参考](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/openedge-function-runtime-python27/openedge_function_runtime_python27.py)

[master.Client](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/module/master/client.go) 可用于调用主程序的 API 来启动、重启或停止临时模块。账号为使用该 Master Client 的常驻模块名，密码使用 `module.GetEnv(module.EnvOpenEdgeModuleToken)` 从环境变量中获取。

[module.Load](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/module/module.go) 可用于加载模块的配置，支持从指定文件中读取 `YAML` 格式的配置，也支持从启动参数中读取 `JSON` 格式的配置。

[module.Wait](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/module/module.go) 可用于等待模块退出的信号。

[logger](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/module/logger/logger.go) 可用于打印日志。
