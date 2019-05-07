# Debian 下 OpenEdge 运行环境配置及快速部署

OpenEdge 主要使用 Go 语言开发，支持两种运行模式，分别是 **docker** 容器模式和 **native** 进程模式。

本文主要介绍 OpenEdge 运行所需环境的安装与配置以及 OpenEdge 在类 Linux 系统下的快速部署。

## 运行环境配置

### Docker 安装

> OpenEdge 提供两种运行方式。如需使用 **docker** 容器模式启动(推荐)，需要先完成 docker 安装。

**提示**：

- 官方提供 Dockerfile 为多阶段镜像构建，如后续需自行构建相关镜像，需要安装 17.05 及以上版本的 docker 来构建 Dockerfile。但生产环境可以使用低版本 docker 来运行镜像，经目前测试，最低可使用版本为 12.0。
- 根据 [官方 Release 日志](https://docs.docker.com/engine/release-notes/#18092) 说明，低于 18.09.2 的 docker 版本具有一些安全隐患，建议安装/更新 docker 版本到 18.09.2 及以上。

可通过以下命令进行安装（适用于类 Linux 系统，[支持多种平台](./Support-platforms.md)）：

```shell
curl -sSL https://get.docker.com | sh
```

**更多内容请参考 [官方文档](https://docs.docker.com/install/)。**

### Python 及其运行时依赖包安装

OpenEdge 提供了 Python 运行时，支持 Python2.7 版本、Python3.6 版本的运行。如计划使用 **native** 进程模式启动，这里推荐本地安装 Python3.6 及运行所依赖的包。如果已经存在 Python3 的其他版本，则可以首先安装 **Python3.6**， 然后使用 `alias` 命令改变系统的默认执行版本为 Python3.6。如果系统的已有程序需要依赖特定的 Python 版本（这里指非 Python3.6），那么用户需要自行保证代码兼容 Python3.6 以保证可以正常使用函数计算服务。如计划以 **docker** 容器模式启动，则无需进行以下步骤。

首先查看 Python3.6 是否已经安装：

```shell
which python3.6
```

如果显示路径的话表明 Python3.6 已经安装成功，不需要进行下面的安装步骤，如果不显示的话表明未安装，则进行下面的安装步骤。Debian 包管理库未直接提供 Python3.6 版本，这里使用源码安装:

```shell
sudo apt-get update
sudo apt-get upgrade
sudo apt-get -y install gcc make zlib1g-dev libffi-dev libssl-dev wget
wget https://www.python.org/ftp/python/3.6.5/Python-3.6.5.tgz
tar -xvf Python-3.6.5.tgz
sudo chmod -R +x Python-3.6.5
cd Python-3.6.5/
./configure
make
sudo make install
sudo pip3 install pyyaml protobuf grpcio
```

等以上命令执行完后输入命令 `python3.6` 确保 Python3.6 安装成功。

### 指定系统默认 Python 版本

用户系统可能存在 Python 的多个版本，所以这里需要指定默认版本为 Python3.6。如不指定，用户需要自行保证编写的代码兼容 Python3.6。

通过以下命令完成(重启有效)：

```shell
alias python=/yourpath/python3.6
```

## OpenEdge 部署

### 部署前准备

**声明**：

- 本文主要介绍 Debian 系统下 OpenEdge 的部署、运行，假定在此之前 OpenEdge [运行所需环境](#运行环境配置) 均已配置完毕。
- 本文所提及的 Debian 系统基于以下内核版本及 CPU 架构，执行命令 `uname -ar` 显示内容如下图所示。

![系统架构及内核版本查询](../../images/setup/os-debian.png)

OpenEdge 容器化模式运行要求运行设备已完成 docker 的安装与运行，可参照 [上述步骤](#docker-安装) 进行安装。

### 部署流程

- Step 1：[下载](../Resources-download.md) OpenEdge 压缩包；
- Step 2：打开终端，进入 OpenEdge 软件包下载目录，进行解压缩操作：
	- 执行命令 `tar -zxvf openedge-xxx.tar.gz`；
- Step 3：完成解压缩操作后，直接进入 OpenEdge 程序包目录，执行命令 `sudo openedge start`，然后分别查看 OpenEdge 启动、加载日志信息，及查看当前正在运行的容器（通过命令 `docker ps`），并对比二者是否一致（假定当前系统中未启动其他 docker 容器）；
- Step 4：若查看结果一致，则表示 OpenEdge 已正常启动。

**注意**：官方下载页面仅提供容器模式程序运行包，如需以进程模式运行，请参考 [源码编译](./Build-OpenEdge-from-Source.md) 相关内容。

### 开始部署

如上所述，首先从 [下载页面](../Resources-download.md) 选择某版本的 OpenEdge 完成下载（也可选择从源码编译，参见 [源码编译](./Build-OpenEdge-from-Source.md)），然后打开终端进入 OpenEdge 程序包下载目录，进行解压缩操作，成功解压缩后，可以发现 openedge 目录中主要包括 bin、etc、var 等目录，具体如下图示。

![OpenEdge 可执行程序包目录](../../images/setup/openedge-dir-debian.png)

其中，`bin` 目录存储 openedge 二进制可执行程序，`etc` 目录存储了 openedge 程序启动的配置，`var` 目录存储了模块启动的配置和资源。

建议把二进制文件放置到 `/usr/local/bin` 或者其他 `PATH` 环境变量中指定的目录中，然后将 `var` 和 `etc` 两个目录拷贝到 `/usr/local` 目录下，或者其他存放二进制文件目录的上级目录中。当然，你也可以将这两个文件夹保留在你解压的位置。

然后，新打开一个终端，执行命令 `docker stats` 查看当前docker中容器的运行状态，如下图示；

![当前运行 docker 容器查询](../../images/setup/docker-stats-before-debian.png)

可以发现，当前系统并未有正在运行的docker容器。

接着，进入解压缩后的 OpenEdge 文件夹下，在另一个终端中执行命令 `sudo openedge start`，如果没有放置 `var` 和 `etc` 目录到二进制文件存放目录的上级目录中，需要通过 `-w` 参数指定工作目录，例： `sudo openedge start -w yourpath/to/configuration`。完成后可通过命令 `ps -ef | grep "openedge"` 来查看运行情况：

![OpenEdge](../../images/setup/openedge-started-thread-debian.png)

另外，可以通过查看日志了解更多 OpenEdge 运行情况。日志的默认存放位置为工作目录下的 `var/log/openedge.log`。

同时观察显示容器运行状态的终端，具体如下图所示：

![当前运行 docker 容器查询](../../images/setup/docker-stats-after-debian.png)

显然，OpenEdge 已经成功启动。

如上所述，若各步骤执行无误，即可完成 OpenEdge 在 Debian 系统上的快速部署、启动。