# Pre-release 0.1.3(2019-05-10)

## 功能

- [#199](https://github.com/baidu/openedge/issues/199) 支持上报服务实例的自定义状态信息，同时采集更多系统状态信息，具体参考 [OpenEdge 系统信息采集](./doc/zh-cn/overview/OpenEdge-design.md#system-inspect)
- [#209](https://github.com/baidu/openedge/issues/209) 当 openedge 启动时，会清理旧实例（残留实例一般由于 openedge 异常退出造成）
- [#211](https://github.com/baidu/openedge/issues/211) 增加了 Python3.6 版本的函数运行时，使用 Ubuntu16.04 作为基础镜像
- [#222](https://github.com/baidu/openedge/issues/222) Docker 容器模式支持 runtime 和 args 配置
- 添加了一个简单的定时器模块 openedge-timer

## Bug 修复

- [#201](https://github.com/baidu/openedge/issues/201) 函数实例池销毁函数实例时，确保停止函数实例
- [#208](https://github.com/baidu/openedge/issues/208) openedge stop 命令等待 openedge 停止运行后退出，确保清理pid文件
- [#234](https://github.com/baidu/openedge/issues/234) hub 模块发布消息等待 ack 超时，快速重发消息
- 解决当 atomic.addUint64() 的参数未按照 64 位对齐导致退出异常的问题。参考：https://github.com/golang/go/issues/23345

## 其他

- [#230](https://github.com/baidu/openedge/issues/230) 发布 OpenEdge 2019 Roadmap
- [#228](https://github.com/baidu/openedge/issues/228) 发布 OpenEdge 社区参与者公约