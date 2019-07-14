# 从源码编译 OpenEdge

相比于快速部署安装 OpenEdge，用户可以采用源码编译的方式来使用 OpenEdge 最新的功能。

在编译源码前，用户应该进行编译环境的配置，所以本文将从 **环境配置** 和 **源码编译** 两方面进行介绍。

## 环境配置

### Linux 平台

#### Go 开发环境安装

前往 [相关资源下载页面](../Resources-download.md) 完成相关二进制包下载。具体请执行：

```shell
tar -C /usr/local -zxf go$VERSION.$OS-$ARCH.tar.gz  # 解压 Go 压缩包至 /usr/local 目录，其中，VERSION、OS、ARCH 参数为下载包对应版本
export PATH=$PATH:/usr/local/go/bin # 设置 Go 执行环境变量
export GOPATH=yourpath  # 设置 GOPATH
go env  # 查看 Go 相关环境变量配置
go version # 查看 Go 版本
```

_**提示**: OpenEdge 要求编译使用的 Go 版本在 1.10.0 以上。_

#### Docker 安装

- 官方提供 Dockerfile 为多阶段镜像构建，如需自行构建相关镜像，需要安装 17.05 及以上版本的 docker 来 build dockerfile。
- 根据[官方 Release 日志](https://docs.docker.com/engine/release-notes/#18092) 说明，低于 18.09.2 的 docker 版本具有一些安全隐患，建议安装/更新 docker 版本到 18.09.2 及以上。

可通过以下命令进行安装（适用于类Linux系统，[支持多种平台](./Support-platforms.md)）：

```shell
curl -sSL https://get.docker.com | sh
```

**Ubuntu**

使用命令

```shell
sudo snap install docker # Ubuntu16.04 after
sudo apt-get install docker.io # Ubuntu 16.04 before
```

即可完成 docker 安装。

**CentOS**

使用命令

```shell
yum install docker
```

即可完成 docker 安装。

**提示**: docker 安装完成后可通过一下命令查看所安装 docker 版本。

```shell
docker version
```

**Debian 9/Raspberry Pi 3**

使用以下命令完成安装：

```shell
curl -sSL https://get.docker.com | sh
```

**更多内容请参考 [官方文档](https://docs.docker.com/install/)。**

### Darwin 平台

#### Go 开发环境配置

+ 通过 HomeBrew 安装

```shell
/usr/bin/ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"  # 安装 HomeBrew
brew install go
```

安装完成后编辑环境配置文件(例: ~/.bash_profile)：

```shell
export GOPATH="${HOME}/go"
export GOROOT="$(brew --prefix golang)/libexec"
export PATH="$PATH:${GOPATH}/bin:${GOROOT}/bin"
```

完成后退出编辑，执行以下命令使环境变量生效：

```shell
source yourfile
```

完成后执行以下命令创建 GOPATH 指定目录:

```shell
test -d "${GOPATH}" || mkdir "${GOPATH}"
```

+ 通过二进制文件安装

前往 [相关资源下载页面](../Resources-download.md) 完成二进制包下载。具体请执行：

```shell
tar -C /usr/local -zxf go$VERSION.$OS-$ARCH.tar.gz  # 解压 Go 压缩包至 /usr/local 目录，其中，VERSION、OS、ARCH 参数为下载包对应版本
export PATH=$PATH:/usr/local/go/bin # 设置 Go 执行环境变量
export GOPATH=yourpath  # 设置 GOPATH
go env  # 查看 Go 相关环境变量配置
go version # 查看 Go 版本
```

_**提示**: OpenEdge 要求编译使用的 Go 版本在 1.10.0 以上。_

#### Docker 安装

前往 [官方页面](https://hub.docker.com/editions/community/docker-ce-desktop-mac) 下载所需 dmg 文件。完成后双击打开，将 docker 拖入 `Application` 文件夹即可。

![Install On Darwin](../../images/setup/docker-install-on-mac.png)

安装完成后使用以下命令查看所安装版本：

```shell
docker version
```

## 源码编译

### 源码下载

按照对应环境完成编译环境配置后，前往 [OpenEdge Github](https://github.com/baidu/openedge) 下载 OpenEdge 源码

```shell
go get github.com/baidu/openedge
```

### 本地镜像构建

容器模式下 docker 通过运行各个模块对应的 **镜像** 来启动该模块，所以首先要构建各个模块的镜像，命令如下：

```shell
cd $GOPATH/src/github.com/baidu/openedge
make rebuild
make image # 在本地生成模块镜像
```

通过上述命令编译生成如下六个镜像:

```shell
openedge-agent:latest
openedge-hub:latest
openedge-function-manager:latest
openedge-remote-mqtt:latest
openedge-function-python27:latest
openedge-function-node85:latest
```

通过以下命令查看生成的镜像：

```shell
docker images
```

### 编译

```shell
cd $GOPATH/src/github.com/baidu/openedge
make rebuild
```

编译完成后会在根目录及各个模块目录下生成如下五个可执行文件:

```shell
openedge
openedge-agent/openedge-agent
openedge-hub/openedge-hub
openedge-function-manager/openedge-function-manager
openedge-remote-mqtt/openedge-remote-mqtt
```

除此之外,会在各个模块目录下共生成五个名为 `package.zip` 文件。

### 安装

默认路径：`/usr/local`。

```shell
cd $GOPATH/src/github.com/baidu/openedge
make install # docker 容器模式安装并使用示例配置
make install-native # native 进程模式安装并使用示例配置
```

指定安装路径，比如安装到 output 目录中。

```shell
cd $GOPATH/src/github.com/baidu/openedge
make install PREFIX=output
```

### 运行

如果程序已经安装到默认路径：`/usr/local`。

```shell
sudo openedge start
```

如果程序已经安装到了指定路径，比如安装到 output 目录中。

```shell
cd $GOPATH/src/github.com/baidu/openedge
sudo ./output/bin/openedge start
```

**注意**：启动方式根据安装方式的不同而不同，即，若选择 docker 运行模式安装，则上述命令会以 docker 容器模式运行 OpenEdge。

**提示**：

- openedge 启动后，可通过 `ps -ef | grep "openedge"` 查看 openedge 是否已经成功运行，并确定启动时所使用的参数。通过查看日志确定更多信息，日志文件默认存放在工作目录下的 `var/log/openedge` 目录中。
- docker 容器模式运行，可通过 `docker stats` 命令查看容器运行状态。
- 如需使用自己的镜像，需要修改应用配置中的模块和函数的 `image`，指定自己的镜像。
- 如需自定义配置，请按照 [配置解读](../tutorials/Config-interpretation.md) 中的内容进行相关设置。

### 卸载

如果是默认安装。

```shell
cd $GOPATH/src/github.com/baidu/openedge
make clean # 可用于清除编译生成的可执行文件
make uninstall # 卸载 docker 容器模式的安装
make uninstall-native # 卸载 native 进程模式的安装
```

如果是指定了安装路径，比如安装到 output 目录中。

```shell
cd $GOPATH/src/github.com/baidu/openedge
make clean # 可用于清除编译生成的可执行文件
make uninstall PREFIX=output # 卸载 docker 容器模式的安装
make uninstall-native PREFIX=output # 卸载 native 进程模式的安装
```
