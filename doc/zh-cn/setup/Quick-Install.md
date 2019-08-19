# 快速安装 OpenEdge

相比于手动下载安装的方式，最新版本 OpenEdge 支持包管理器的安装方式。通过包管理器方式，用户可以在终端简单输入几条命令，快速安装最新版本的 OpenEdge。

OpenEdge 包安装器目前支持的系统有: Ubuntu Server 16.04、Ubuntu Server 18.04、Debian9、CentOS7、Raspbian-stretch，支持的平台有 X64、i386、ARM32 和 ARM64。

OpenEdge 支持两种运行模式，分别是 docker 容器模式和 native 进程模式。本文使用 **docker** 容器模式进行快速安装。

## 安装容器运行时

在 **docker** 容器模式下，OpenEdge 依赖于 docker 容器运行时。如果用户机器尚未安装 docker 容器运行时，可通过以下命令来安装 docker 的最新版本（适用于类 Linux 系统）:

```shell
curl -sSL https://get.docker.com | sh
```

安装结束后，可以查看 docker 的版本号:

```shell
docker version
```

**注意**：根据 官方 Release 日志 说明，低于 18.09.2 的 Docker 版本具有一些安全隐患，建议安装/更新 Docker 版本到 18.09.2 及以上。

## 安装 OpenEdge

在 OpenEdge 发布新版本的同时，也会发布对应的 rpm、deb 包。使用以下命令，用户可以通过包管理方式将 OpenEdge 的最新版本安装到设备上:

```shell
curl -sSL http://download.openedge.tech/install.sh | sudo -E bash -
```

最新版本 OpenEdge 支持 Systemd 守护，可以使用以下命令开启守护:

```shell
sudo systemctl start openedge
```

## 导入默认配置包（可选）

OpenEdge 作为一个边缘计算平台，除了提供底层服务管理能力外，还提供一些基础功能模块。用户可以通过配置文件，加载相应的模块以及设定模块本身的运行参数。OpenEdge 官方提供了一套默认配置，可以通过以下命令导入默认配置文件:

```shell
curl -sSL http://download.openedge.tech/install_with_docker_example.sh | sudo -E bash -
```

默认配置只用于学习和测试目的，用户应根据自己的实际工作场景进行按需配置，具体可参考 [配置解读](../tutorials/Config-interpretation.md) 中的内容进行进一步了解。

如果用户不需要启动任何模块，那就不需要导入任何配置文件。

## 验证是否成功安装

- 在终端中命令 `sudo systemctl status openedge` 来查看 `openedge` 是否正常运行。正常如下图所示，否则说明主程序 `openedge` 启动失败；

![OpenEdge](../../images/setup/openedge-systemctl-status.png)

- 在终端中执行命令 `docker stats` 查看当前 docker 中容器的运行状态。由于主程序 `openedge` 会先到镜像仓库拉取需要的镜像，用户需要等待 3~5 分钟执行此条命令。以上一步中导入的默认配置为例，待主程序拉取完成后，容器的运行状态如下图所示。如果用户本地的镜像与下述不一致，说明模块启动失败；

![当前运行 docker 容器查询](../../images/setup/docker-stats.png)

- 针对上述两种失败情况，用户需要查看日志来了解具体的错误情况。日志的默认存放位置为 `/usr/local/var/log/openedge/openedge.log`。针对日志中出现的错误，用户可先参考常见问题 [常见问题](../FAQ.md) 进行解决。必要时可以直接[提交 Issue](https://github.com/baidu/openedge/issues)。
