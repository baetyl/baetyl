# Mac下OpenEdge安装及环境配置

> OpenEdge 主要使用 Go 语言开发，支持两种运行模式，分别是 ***docker*** 容器模式和 ***native*** 进程模式。本文主要介绍 OpenEdge 程序的安装以及运行所需环境的安装与配置。

## OpenEdge 程序安装

前往[下载页面](https://github.com/baidu/openedge/releases)找到机器对应版本并进行下载，推荐下载最新版程序运行包。

***注：*** 官方下载页面仅提供容器模式程序运行包，如需以线程模式运行，请参考从[源码编译](./Source.md)相关内容。

## 运行环境配置

### Go 开发环境安装

前往[下载页面](https://golang.org/dl/)完成相关包下载。或使用命令，如：

```shell
wget https://dl.google.com/go/go1.11.2.linux-amd64.tar.gz
```
获取最新安装包，其中OpenEdge程序要求***Go语言版本***不低于 **1.10.0**。

解压下载的安装包到本地用户文件夹。

```shell
tar -C /usr/local -zxf go$VERSION.$OS-$ARCH.tar.gz
```

其中，VERSION、OS、ARCH参数为下载包对应版本。

导入环境变量：

```shell
export PATH=$PATH:/usr/local/go/bin
```

完成后通过以下命令查看版本:

```shell
go version
```

或通过以下命令查看go相关环境配置：

```shell
go env
```

更多请参考[官方文档](https://golang.org/doc/install)。

### Docker 安装

前往[官方页面](https://hub.docker.com/editions/community/docker-ce-desktop-mac)下载所需 dmg 文件。完成后双击打开，将 Docker 拖入 Application 文件夹即可。

![Install On Mac](../images/install/docker_install_on_mac.png)

安装完成后使用以下命令查看所安装版本：

```shell
docker version
```

***注：*** 官方提供 Dockerfile 为多阶段镜像构建，如需自行构建相关镜像，建议 Docker 版本在 17.05 之上。

**更多内容请参考[官方文档](https://docs.docker.com/install/)。**

### Python 安装

推荐使用 HomeBrew 安装。

```shell
/usr/bin/ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"  // 安装HomeBrew
brew install python@2
```

***注*** : 安装完成后可通过以下命令查看所安装版本：

```shell
python -V
```

通过以下命令设置默认 Python 命令指定上述安装的版本。例如：

```shell
alias python=/yourpath/python2.7
```

### Python Runtime 依赖模块安装

按照上述步骤完成 Python 2.7版本的安装后，需要安装 Python Runtime 运行所需模块：

```shell
pip install pyyaml protobuf grpcio
```

