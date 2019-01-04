# 从源码编译 OpenEdge

## Linux下编译环境配置

### Go 开发环境安装

前往[相关资源下载页面](../Resources-download.md)完成相关二进制包下载。具体请执行：

```shell
tar -C /usr/local -zxf go$VERSION.$OS-$ARCH.tar.gz  # 解压Go压缩包至/usr/local目录，其中，VERSION、OS、ARCH参数为下载包对应版本
export PATH=$PATH:/usr/local/go/bin # 设置Go执行环境变量
export GOPATH=yourpath  # 设置GOPATH
go env  # 查看Go相关环境变量配置
go version # 查看Go版本
```

_**提示**: OpenEdge要求编译使用的Go版本在 1.10.0 以上。_

### Docker 安装

> 官方提供 Dockerfile 为多阶段镜像构建，如需自行构建相关镜像，需要安装17.05 及以上版本的 Docker 来build Dockerfile。

可通过以下命令进行安装（适用于类Linux系统，[支持多种平台](./Support-platforms.md)）：

```shell
curl -sSL https://get.docker.com | sh
```

#### Ubuntu

使用命令

```shell
sudo snap install docker # Ubuntu16.04 after
sudo apt-get install docker.io # Ubuntu 16.04 before
```

即可完成 Docker 安装。

#### CentOS

使用命令

```shell
yum install docker
```

即可完成 docker 安装。

***注*** : Docker 安装完成后可通过一下命令查看所安装Docker版本。

```shell
docker version
```

#### Debian 9/Raspberry Pi 3

使用以下命令完成安装：

```shell
curl -sSL https://get.docker.com | sh
```

**更多内容请参考[官方文档](https://docs.docker.com/install/)。**

## Darwin下编译环境配置

### Go 开发环境配置

+ 通过 HomeBrew 安装

```shell
/usr/bin/ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"  # 安装HomeBrew
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

完成后执行以下命令创建GOPATH指定目录:

```shell
test -d "${GOPATH}" || mkdir "${GOPATH}"
```

+ 通过二进制文件安装

前往[相关资源下载页面](../Resources-download.md)完成二进制包下载。具体请执行：

```shell
tar -C /usr/local -zxf go$VERSION.$OS-$ARCH.tar.gz  # 解压Go压缩包至/usr/local目录，其中，VERSION、OS、ARCH参数为下载包对应版本
export PATH=$PATH:/usr/local/go/bin # 设置Go执行环境变量
export GOPATH=yourpath  # 设置GOPATH
go env  # 查看Go相关环境变量配置
go version # 查看Go版本
```

_**提示**: OpenEdge要求编译使用的Go版本在 1.10.0 以上。_

### Docker 安装

前往[官方页面](https://hub.docker.com/editions/community/docker-ce-desktop-mac)下载所需 dmg 文件。完成后双击打开，将 Docker 拖入 Application 文件夹即可。

![Install On Darwin](../../images/setup/docker_install_on_mac.png)

安装完成后使用以下命令查看所安装版本：

```shell
docker version
```

## 源码下载

按照对应环境完成编译环境配置后，前往[OpenEdge Github](https://github.com/baidu/openedge)下载openedge源码

```shell
go get github.com/baidu/openedge
```

## 源码编译

```shell
cd $GOPATH/src/github.com/baidu/openedge
make # 编译主程序和模块的可执行程序
```

编译完成后会在根目录下生成如下四个可执行文件:

```shell
openedge
openedge-hub
openedge-function
openedge-remote-mqtt
```

```shell
make images # 在本地生成模块镜像
```

通过上述命令编译生成如下四个镜像:

```shell
openedge-hub:build
openedge-function:build
openedge-remote-mqtt:build
openedge-function-runtime-python27:build
```

通过以下命令查看生成的镜像：

```shell
docker images
```

## 程序安装

装到默认路径：/usr/local。

```shell
cd $GOPATH/src/github.com/baidu/openedge
make install
```

指定安装路径，比如安装到output目录中。

```shell
cd $GOPATH/src/github.com/baidu/openedge
make PREFIX=output install
```

## 程序运行

如果程序已经安装到默认路径：/usr/local。

```shell
openedge -w example/docker # docker容器模式
openedge # native进程模式
```

如果程序已经安装到了指定路径，比如安装到output目录中。

```shell
cd $GOPATH/src/github.com/baidu/openedge
output/bin/openedge -w example/docker # docker容器模式
output/bin/openedge # native进程模式
```

**提示**：

1. docker容器模式运行，可通过 ```docker stats``` 命令查看容器运行状态。
2. 如需使用自己的镜像，需要修改应用配置中的模块和函数的 entry，指定自己的镜像。
3. 如需自定义配置，请按照 [配置解读](../tutorials/local/Config-interpretation.md) 中的内容进行相关设置。

## 程序卸载

如果是默认安装。

```shell
cd $GOPATH/src/github.com/baidu/openedge
make clean # 可用于清除编译生成的可执行文件
make uninstall
```

如果是指定了安装路径，比如安装到output目录中。

```shell
cd $GOPATH/src/github.com/baidu/openedge
make clean # 可用于清除编译生成的可执行文件
make PREFIX=output uninstall
```