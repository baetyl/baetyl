# Pre-release 0.1.2(2019-04-04)

## Features

> + 支持 Linux-arm64 平台；
> + 支持从 `hub.docker.com` 直接拉取镜像，方便国内外用户使用；
> + 针对资源配置引入数据卷（volume）概念，修改一些语义不准确配置项，提供更灵活、规范、准确的配置；
> + 从主程序分离 Agent 模块，定时上报（核心）设备状态信息，具体可参考 [issue 20](https://github.com/baidu/openedge/issues/20)；
> + OpenEdge 支持以命令行发布，启动，后台服务运行，具体可参考 [issue 122](https://github.com/baidu/openedge/issues/122)。

## Bug fixes

> + 修复一些文档描述错误；
> + 针对容器模式，修复 Linux 主程序服务 socket 地址不可用 bug，具体可参考 [issue 92](https://github.com/baidu/openedge/issues/92)。

## Others(include release engineering)

> + 更丰富的测试 example 支持，如针对 Hub 接入模块，提供基于 mosquitto 的接入测试配置；
> + 全量、更丰富的英文文档支持（OpenEdge 概述、安装、部署、使用、开发自定义模块等），具体可参考 [issue 61](https://github.com/baidu/openedge/issues/61)