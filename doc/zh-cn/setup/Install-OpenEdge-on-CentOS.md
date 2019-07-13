# CentOS 下 OpenEdge 运行环境配置及快速部署

OpenEdge 主要使用 Go 语言开发，支持两种运行模式，分别是 **docker** 容器模式和 **native** 进程模式。

本文主要介绍 OpenEdge 运行所需环境的安装以及 OpenEdge 在类 Linux 系统下的快速部署。

**声明**：

- 本文测试系统基于 CentOS7.5 x86_64 版本，内核及CPU架构信息通过执行 `uname -ar` 命令查看如下：
![系统架构及内核版本查询](../../images/setup/os-centos.png)
- 在 OpenEdge 部署小节中，使用 **docker** 容器模式演示部署流程。

## 运行环境配置

OpenEdge 提供 **docker** 容器模式和 **native** 进程模式。如果以 **docker** 容器模式运行，需要安装 Docker 环境；如果以 **native** 进程模式运行，需要安装 Python 及其运行时依赖包。

### 容器模式下安装 Docker

如需使用 **docker** 容器模式启动(推荐)，需要先完成 Docker 安装。

**提示**：

- 官方提供 Dockerfile 为多阶段镜像构建，如后续需自行构建相关镜像，需要安装 17.05 及以上版本的 Docker 来构建 Dockerfile。但生产环境可以使用低版本 Docker 来运行镜像，经目前测试，最低可使用版本为 12.0。
- 根据 [官方 Release 日志](https://docs.docker.com/engine/release-notes/#18092) 说明，低于 18.09.2 的 Docker 版本具有一些安全隐患，建议安装/更新 Docker 版本到 18.09.2 及以上。

可通过以下命令进行安装（适用于类 Linux 系统，[支持多种平台](./Support-platforms.md)）：

```shell
curl -sSL https://get.docker.com | sh
```

或者使用命令

```shell
sudo yum install docker
```

即可完成 Docker 安装。

***注意*** :

+ Docker 安装完成后可通过一下命令查看所安装 Docker 版本。

```shell
docker version
```

**更多内容请参考 [官方文档](https://docs.docker.com/install/)。**

### 进程模式下安装依赖

OpenEdge 提供了 Python 运行时、Node 运行时。如计划使用 **native** 进程模式启动，需要本地安装这些运行时环境及其相关依赖。对应版本分别为 Python2.7、Python3.6、Node8.5。用户也可以选择其他版本，但需要自行保证兼容性。

#### 安装 Python 运行时

系统默认提供 Python2.7，接下里介绍 Python3.6 安装过程。

- Step 1：查看是否已经安装 Python3.6 或以上版本。如果是则直接执行 Step 3，否则执行 Step 2。

```shell
which python3
```

- Step 2：安装 Python3.6:

```shell
sudo yum install -y epel-release
sudo yum update
sudo yum install -y python36
```

- Step 3：安装 OpenEdge 需要的相关依赖包:

```shell
# python2
sudo yum install -y python2-pip
sudo pip2 install grpcio protobuf pyyaml
sudo pip2 install -U PyYAML

# python3
sudo yum install -y python36-pip
sudo pip3 install grpcio protobuf pyyaml
sudo pip3 install -U PyYAML
```

#### 安装 Node 运行时

- Step 1：查看是否已经安装 Node8.5 或以上版本。如果没有则执行 Step 2。

```shell
node -v
```

- Step 2：安装 Node8:

```shell
curl -sL https://rpm.nodesource.com/setup_8.x | bash -
sudo yum install -y nodejs
```

## OpenEdge 部署

### 部署流程

- Step 1：下载 [OpenEdge](../Resources-download.md) 压缩包；
- Step 2：打开终端，进入 OpenEdge 软件包下载目录，进行解压缩操作：

```shell
unzip openedge-xxx.zip
```

- Step 3：完成解压缩操作后，进入 OpenEdge 程序包目录，执行命令 `sudo openedge start` 启动 OpenEdge。随后查看 OpenEdge 的启动、加载日志，并使用 `docker ps` 命令查看此时正在运行的 docker 容器，分析两者检查 OpenEdge 需要启动的镜像是否全部加载成功；
- Step 4：如果日志中要启动的镜像全部在 docker 中通过容器加载成功，则表示 OpenEdge 启动成功。

**注意**：官方下载页面仅提供容器模式程序运行包，如需以进程模式运行，请参考 [源码编译](./Build-OpenEdge-from-Source.md) 相关内容。

### 开始部署

如上所述，首先从 [下载页面](../Resources-download.md) 选择某版本的 OpenEdge 完成下载（也可选择从源码编译，参见 [源码编译](./Build-OpenEdge-from-Source.md)），然后打开终端进入 OpenEdge 程序包下载目录，进行解压缩操作，成功解压缩后，可以发现 openedge 目录中主要包括 bin、etc、var 等目录，具体如下图示。

![OpenEdge 可执行程序包目录](../../images/setup/openedge-dir-centos.png)

其中，`bin` 目录存储 `openedge` 二进制可执行程序，`etc` 目录存储了程序启动的配置，`var` 目录存储了模块启动的配置和资源。

建议把二进制文件放置到 `/usr/local/bin` 或者其他 `PATH` 环境变量中指定的目录中，然后将 `var` 和 `etc` 两个目录拷贝到 `/usr/local` 目录下，或者其他存放二进制文件目录的上级目录中。当然，你也可以将这两个文件夹保留在你解压的位置。

然后，新打开一个终端，执行命令 `docker stats` 查看当前 docker 中容器的运行状态，如下图示；

![当前运行 docker 容器查询](../../images/setup/docker-stats-before-centos.png)

可以发现，当前系统并未有正在运行的 docker 容器。

接着，进入解压缩后的 OpenEdge 文件夹下，在另一个终端中执行命令 `sudo openedge start`，如果没有放置 `var` 和 `etc` 目录到二进制文件存放目录的上级目录中，需要通过 `-w` 参数指定工作目录，例： `sudo openedge start -w yourpath/to/configuration`。完成后可通过命令 `ps -ef | grep "openedge"` 来查看运行情况：

![OpenEdge](../../images/setup/openedge-started-thread-centos.png)

另外，可以通过查看日志了解更多 OpenEdge 运行情况。日志的默认存放位置为工作目录下的 `var/log/openedge.log`。

同时观察显示容器运行状态的终端，具体如下图所示：

![当前运行 docker 容器查询](../../images/setup/docker-stats-after-centos.png)

显然，OpenEdge 已经成功启动。

如上所述，若各步骤执行无误，即可完成 OpenEdge 在 CentOS 系统上的快速部署、启动。