# Linux下OpenEdge安装及环境配置

> OpenEdge 主要使用 Go 语言开发，支持两种运行模式，分别是 ***docker*** 容器模式和 ***native*** 进程模式。

本文主要介绍 OpenEdge 程序的安装以及运行所需环境的安装与配置。

## OpenEdge 程序安装

前往[下载页面](https://github.com/baidu/openedge/releases)找到机器对应版本并进行下载，完成后解压到任意目录。(推荐下载最新版程序运行包)

***注：*** 官方下载页面仅提供容器模式程序运行包，如需以进程模式运行，请参考从[源码编译](./Build-OpenEdge-from-Source.md)相关内容。

## 运行环境配置

### Docker 安装

> OpenEdge 提供两种运行方式。如需使用 ***docker*** 容器模式启动(推荐)，需要先完成 docker 安装。

可通过以下命令进行安装（适用于类Linux系统，[支持多种平台](./Support-platforms.md)）：

```shell
curl -sSL https://get.docker.com | sh
```

#### Ubuntu

使用命令

```shell
sudo snap install docker # Ubuntu16.04 after
sudo apt install docker.io # Ubuntu 16.04 before
```

即可完成 Docker 安装。

#### CentOS

使用命令

```shell
yum install docker
```

即可完成 docker 安装。

***注意*** : 

+ Docker 安装完成后可通过一下命令查看所安装 Docker 版本。

```shell
docker version
```

+ 官方提供 Dockerfile 为多阶段镜像构建，如后续需自行构建相关镜像，需要安装17.05 及以上版本的 Docker 来build Dockerfile。但生产环境可以使用低版本 Docker 来运行镜像，经目前测试，最低可使用版本为 12.0。

#### Debian 9/Raspberry Pi 3

使用以下命令完成安装：

```shell
curl -sSL https://get.docker.com | sh
```

**更多内容请参考[官方文档](https://docs.docker.com/install/)。**

### Python2.7 及 Python Runtime 依赖包安装

> + OpenEdge 提供了 Python Runtime，支持 Python 2.7 版本的运行，如计划使用 ***native*** 进程模式启动，需要安装 Python 2.7 及运行所依赖的模块。如计划以 ***docker*** 容器模式启动，则无需进行以下步骤。

#### Ubuntu 18.04 LTS/Debian 9/Raspberry Pi 3

使用如下命令安装 Python 2.7:

```shell
sudo apt update
sudo apt upgrade
sudo apt install python2.7
sudo apt install python-pip
sudo pip install protobuf grpcio
```

#### CentOs 

执行以下命令检查已安装Python版本：

```shell
python -V
```

如果显示未安装，可使用以下命令进行安装：

```shell
yum install python
yum install python-pip
yum install protobuf grpcio
```

或者通过源码编译安装：

```shell
yum install gcc openssl-devel bzip2-devel
wget https://www.python.org/ftp/python/2.7.15/Python-2.7.15.tgz
tar xzf Python-2.7.15.tgz
make altinstall
curl "https://bootstrap.pypa.io/get-pip.py" -o "get-pip.py"
python2.7 get-pip.py
pip install protobuf grpcio
```

输入命令查看 Python 版本为 2.7.* 后为安装正确。

### 指定默认 Python 版本

某些情况下需要指定默认 Python 版本为上述安装版本。通过以下命令完成(重启有效)：

```shell
alias python=/yourpath/python2.7
```

## 常见问题

A. ***Got permission denied while trying to connect to the Docker daemon socket at unix:///var/run/docker.sock: Get http://%2Fvar%2Frun%2Fdocker.sock/v1.38/images/json: dial unix /var/run/docker.sock: connect: permission denied***

1. 提供管理员权限
2. 通过以下命令添加当前用户到docker用户组：

```shell
sudo usermod -aG docker ${USER}
su - ${USER}
``` 

如提示没有 docker group，使用如下命令创建新docker用户组后再执行上述命令：

```shell
sudo groupadd docker
```

B. ***Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?***

按照问题A解决方案执行后如仍报出此问题，重新启动docker服务即可。

例，CentOs 下启动命令：

```shell
systemctl start docker
```