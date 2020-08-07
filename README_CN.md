# BAETYL v2

[![Baetyl-logo](./docs/logo_with_name.png)](https://baetyl.io)

[![build](https://github.com/baetyl/baetyl/workflows/build/badge.svg)](https://github.com/baetyl/baetyl/actions?query=workflow%3Abuild)
[![codecov](https://codecov.io/gh/baetyl/baetyl/branch/master/graph/badge.svg)](https://codecov.io/gh/baetyl/baetyl)
[![Go Report Card](https://goreportcard.com/badge/github.com/baetyl/baetyl)](https://goreportcard.com/report/github.com/baetyl/baetyl) 
[![License](https://img.shields.io/github/license/baetyl/baetyl?color=blue)](LICENSE) 
[![Stars](https://img.shields.io/github/stars/baetyl/baetyl?style=social)](Stars)

[![README](https://img.shields.io/badge/README-English-brightgreen)](./README.md)

**[Baetyl](https://baetyl.io) 是 [Linux Foundation Edge](https://www.lfedge.org) 
旗下的边缘计算项目，旨在将云计算能力拓展至用户现场**。
提供临时离线、低延时的计算服务，包括设备接入、消息路由、数据遥传、函数计算、视频采集、AI推断、状态上报、配置下发等功能。

Baetyl v2 提供了一个全新的边云融合平台，采用云端管理、边缘运行的方案，分成[**边缘计算框架（本项目）**](https://github.com/baetyl/baetyl)和[**云端管理套件**](https://github.com/baetyl/baetyl-cloud)两部分，支持多种部署方式。可在云端管理所有资源，比如节点、应用、配置等，自动部署应用到边缘节点，满足各种边缘计算场景，特别适合新兴的强边缘设备，比如 AI 一体机、5G 路侧盒子等。

v2 和 v1 版本的主要区别如下：
* 边缘和云端框架全部向云原生演化，已支持运行在 K8S 或 K3S 之上。
* 引入声明式的设计，通过影子（Report/Desire）实现端云同步（OTA）。
* 边缘框架目前支持 Kube 模式（Kube Mode），由于运行在 K3S 上，整体的资源开销较大（1G内存）；进程模式（Native Mode）正在开发中，可以大大降低资源消耗。
* 边缘框架将来会支持边缘节点集群。

## 架构

![Architecture](./docs/baetyl-arch-v2.svg)

### [边缘计算框架（本项目）](./README_CN.md)

边缘计算框架（Edge Computing Framework）运行在边缘节点的 Kubernetes 上，
管理和部署节点的所有应用，通过应用服务提供各式各样的能力。
应用包含系统应用和普通应用，系统应用全部由 Baetyl 官方提供，用户无需配置。

目前有如下几个系统应用：
* baetyl-init：负责激活边缘节点到云端，并初始化 baetyl-core，任务完成后就会退出。
* baetyl-core：负责本地节点管理（node）、端云数据同步（sync）和应用部署（engine）。
* baetyl-function: 所有函数运行时服务的代理模块，函数调用都到通过这个模块。

目前框架支持 Linux/amd64、Linux/arm64、Linux/armv7，
如果边缘节点的资源有限，可考虑使用轻量版 Kubernetes：[K3S](https://k3s.io/)。

边缘节点的硬件要求取决于你要部署的应用，推荐的最低要求如下：
* 内存 1GB
* CPU 1核

### [云端管理套件](https://github.com/baetyl/baetyl-cloud)

云端管理套件（Cloud Management Suite）负责管理所有资源，包括节点、应用、配置、部署等。所有功能的实现都插件化，方便功能扩展和第三方服务的接入，提供丰富的应用。云端管理套件的部署非常灵活，即可部署在公有云上，又可部署在私有化环境中，还可部署在普通设备上，支持 K8S/K3S 部署，支持单租户和多租户。

开源版云端管理套件提供的基础功能如下：
* 边缘节点管理
    * 在线安装
    * 端云同步（影子）
    * 节点信息
    * 节点状态
    * 应用状态
* 应用部署管理
    * 容器应用
    * 函数应用
    * 节点匹配（自动）
* 配置管理
    * 普通配置
    * 函数配置
    * 密文
    * 证书
    * 镜像库凭证
* 节点预配管理
    * 批次管理
    * 注册激活

_开源版本包含上述所有功能的 RESTful API，暂不包含前端界面（Dashboard）。_

## 联系我们

Baetyl 作为中国首发的开源边缘计算框架，
我们旨在打造一个 **轻量、安全、可靠、可扩展性强** 的边缘计算社区，
为边缘计算技术的发展和不断推进营造一个良好的生态环境。
为了更好的推进 Baetyl 的发展，如果您有更好的关于 Baetyl 的发展建议，
欢迎选择如下方式与我们联系。

- 欢迎加入 [Baetyl 边缘计算开发者社区群](https://baetyl.bj.bcebos.com/Wechat/Wechat-Baetyl.png)
- 欢迎加入 [Baetyl 的 LF Edge 讨论组](https://lists.lfedge.org/g/baetyl/topics)
- 欢迎发送邮件到：<baetyl@lists.lfedge.org>
- 欢迎到 [GitHub 提交 Issue](https://github.com/baetyl/baetyl/issues)

## 如何贡献

如果您热衷于开源社区贡献，Baetyl 将为您提供两种贡献方式，分别是代码贡献和文档贡献。
具体请参考 [如何向 Baetyl 贡献代码和文档](./docs/contributing_cn.md)。
