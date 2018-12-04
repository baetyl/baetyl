# 环境配置

- [go开发环境安装](#go开发环境安装)
  - [Linux环境下安装](#Linux环境下安装)
- [godep环境配置](#godep环境配置)
- [docker安装](#docker安装)
  - [Linux环境下安装](#Linux环境下安装)

> OpenEdge使用Go语言编写，使用Godep工具管理相关依赖，支持两种运行模式， 分别是**docker**容器模式和**native**进程模式。此文介绍相关环境的安装及配置。

## go开发环境安装

### Linux环境下安装

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

最后，导入相应环境变量即可：

```shell
export PATH=$PATH:/usr/local/go/bin
```

更多请参考[官方文档](https://golang.org/doc/install)。

## godep环境配置

OpenEdge使用 [GoDep工具](https://github.com/tools/godep#godep---archived) 来管理相关依赖。
在安装配置完go语言开发环境后，只需执行以下命令即可完成安装。

```shell
go get github.com/tools/godep
```

## docker安装

### Linux环境

本文主要介绍在Ubuntu Xenial 16.04 (LTS)版本下的安装，更多请参考[官方文档](https://docs.docker.com/install/)。

1.移除旧版本docker

```shell
sudo apt-get remove docker docker-engine docker.io
```

2.更新包索引

```shell
sudo apt-get update
```

3.安装https相关包

```shell
sudo apt-get install \
    apt-transport-https \
    ca-certificates \
    curl \
    software-properties-common
```

4.添加docker官方GPG秘钥

```shell
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
```

5.设置库房

**x86_64/amd64环境下执行：**

```shell
sudo add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
   $(lsb_release -cs) \
   stable"
```

6.安装最新版本docker

```shell
sudo apt-get install docker-ce
```

**注** : 默认情况下，ubuntu需要sudo权限才可以执行docker相关操作。可以通过下述命令进行修改：

```shell
sudo usermod -aG docker ${USER}
su - ${USER}
```

通过命令查看权限是否添加成功。

```shell
id -nG
```
