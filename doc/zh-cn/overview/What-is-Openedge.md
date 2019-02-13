# 什么是 OpenEdge

[OpenEdge](https://openedge.tech) 是百度云开源的边缘计算相关产品，可将云计算能力拓展至用户现场，提供临时离线、低延时的计算服务，包括设备接入、消息路由、消息远程同步、函数计算等功能。OpenEdge 和[智能边缘 BIE](https://cloud.baidu.com/product/bie.html)（Baidu-IntelliEdge）云端管理套件配合使用，通过在云端进行智能边缘核心设备的建立、身份制定、策略规则制定、函数编写，然后生成配置文件下发至 OpenEdge 本地运行包，可达到云端管理和应用下发，边缘设备上运行应用的效果，满足各种边缘计算场景。

在[架构设计](./OpenEdge-design.md)上，OpenEdge 一方面推行 **模块化**，拆分各项主要功能，确保每一项功能都是一个独立的模块，整体由主程序模块控制启动、退出，确保各项子功能模块运行互不依赖、互不影响；总体上来说，推行模块化的设计模式，可以充分满足用户 **按需使用、按需部署** 的切实要求；另一方面，OpenEdge 在设计上还采用 **容器化** 的设计思路，基于各模块的镜像可以在 Docker 支持的各类操作系统上进行 **一键式构建**，依托 Docker 跨平台支持的特性，确保 OpenEdge 在各系统、各平台的环境一致、标准化；此外，OpenEdge 还针对 Docker 容器化模式赋予其 **资源隔离与限制** 能力，精确分配各运行实例的 CPU、内存等资源，提升资源利用效率。

更多 OpenEdge 构成、功能、特性与优势等内容，请参考：

> + [OpenEdge 设计](./OpenEdge-design.md)
> + [OpenEdge 构成](./OpenEdge-concepts.md)
> + [OpenEdge 功能](./OpenEdge-features.md)
> + [OpenEdge 优势](./OpenEdge-advantages.md)
> + [OpenEdge 开放式设计框架](./OpenEdge-open-framework.md)
> + [OpenEdge 自定义扩展](./OpenEdge-extension.md)
> + [OpenEdge 安全](./OpenEdge-security.md)
> + [OpenEdge 可控式设计](./OpenEdge-control.md)