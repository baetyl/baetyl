# Pre-release 0.1.3(2019-05-10)

## Features

- 支持上报服务实例的自定义状态信息，同时采集更多系统状态信息，具体参考：[https://github.com/baidu/openedge/blob/master/doc/zh-cn/overview/OpenEdge-design.md#system-inspect](https://github.com/baidu/openedge/blob/master/doc/zh-cn/overview/OpenEdge-design.md#system-inspect)
- 当 openedge 启动时，会清理旧实例（残留实例一般由于 openedge 异常退出造成）
- 增加了 Python3.6 版本的函数运行时，使用 Ubuntu16.04 作为基础镜像
- Docker 容器模式支持 runtime 和 args 配置
添加了一个简单的定时器模块 openedge-timer

## Bug Fixs

- 函数实例池销毁函数实例时，确保停止函数实例
- openedge stop 命令等待 openedge 停止运行后退出，确保清理pid文件
- hub 模块发布消息等待 ack 超时，快速重发消息
- 解决当 atomic.addUint64() 的参数未按照 64 位对齐导致退出异常的问题。参考：[https://github.com/golang/go/issues/23345](https://github.com/golang/go/issues/23345)

## Others(include release engineering)

- 发布 OpenEdge 2019 Roadmap
- 发布 OpenEdge 社区参与者公约