# Darwin下OpenEdge安装及环境配置

> OpenEdge 主要使用 Go 语言开发，支持两种运行模式，分别是 ***Docker*** 容器模式和 ***Native*** 进程模式。

本文主要介绍 OpenEdge 程序的安装以及运行所需环境的安装与配置。

## OpenEdge 程序安装

前往[下载页面](https://github.com/baidu/openedge/releases)找到机器对应版本并进行下载，完成后解压到任意目录。(推荐下载最新版程序运行包)

***注：*** 官方下载页面仅提供容器模式程序运行包，如需以进程模式运行，请参考从[源码编译](./Build-OpenEdge-from-Source.md)相关内容。

## 运行环境配置

### Docker 安装

> OpenEdge 提供两种运行方式。如需使用 ***Docker*** 容器模式启动(推荐)，需要先完成 Docker 安装。

前往[官方页面](https://hub.docker.com/editions/community/docker-ce-desktop-mac)下载所需 dmg 文件。完成后双击打开，将 Docker 拖入 Application 文件夹即可。

![Install On Darwin](../../images/setup/docker_install_on_mac.png)

安装完成后使用以下命令查看所安装版本：

```shell
docker version
```

***注：*** 官方提供 Dockerfile 为多阶段镜像构建，如需自行构建相关镜像，需要安装17.05 及以上版本的 Docker 来build Dockerfile。但生产环境可以使用低版本 Docker 来运行镜像，经目前测试，最低可使用版本为 12.0。

**更多内容请参考[官方文档](https://docs.docker.com/install/)。**

### Python2.7安装及 Python Runtime 运行所依赖模块安装

> + OpenEdge 提供了 Python Runtime，支持 Python 2.7 版本的运行，如计划使用 ***Native*** 进程模式启动，需要安装 Python 2.7 及运行所依赖的模块。如计划以 ***Docker*** 容器模式启动，则无需进行以下步骤。

推荐使用 HomeBrew 安装。

```shell
/usr/bin/ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"  // 安装HomeBrew
brew install python@2
pip isntall protobuf grpcio
```

***注*** : 安装完成后可通过以下命令查看所安装版本：

```shell
python -V
```

通过以下命令设置默认 Python 命令指定上述安装的版本。例如：

```shell
alias python=/yourpath/python2.7
```