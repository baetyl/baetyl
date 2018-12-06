本文主要提供OpenEdge在各系统、平台部署、启动的相关问题及解决方案。

**问题1**：在以容器模式启动OpenEdge时，提示docker版本不正确，具体如下图示。

![图片](../../images/develop/start/macos/docker-api-version-faq.png)

**参考方案**：如上图所示，容器模式启动OpenEdge时，提示docker client版本不正确（过高），与docker API version不兼容，可以通过设置DOCKER_API_VERSION环境变量解决（`export DOCKER_API_VERSION=1.39`，这里设置不高于1.39即可）。

**问题2**：在以容器模式启动OpenEdge时，提示缺少启动依赖配置项。

![图片](../../images/develop/start/macos/docker-engine-conf-miss.png)

**参考方案**：如上图所示，OpenEdge启动缺少配置依赖文件，参考[OpenEdge设计文档](../../about/OpenEdge整体设计.md)及[GitHub项目开源包](https://github.com/baidu/openedge)example文件夹补充相应配置文件即可。