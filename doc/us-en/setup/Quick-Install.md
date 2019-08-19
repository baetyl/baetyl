# Quick Install OpenEdge

Compared to the previous version which using manual download installation, the latest version (version 0.1.5) OpenEdge supports package manager installation. With package manager, users can quickly install OpenEdge by simply typing a few commands at the terminal.

The OpenEdge package manager currently supports the following systems: Ubuntu Server 16.04, Ubuntu Server 18.04, Debian9, CentOS7, and Raspbian-stretch. The supported platforms are X64, i386, ARM32, and ARM64.

OpenEdge supports two running modes: **docker** container mode and **native** process mode. This article's quick installation is based on **docker** container mode.

## Install the container runtime

In **docker** container mode, OpenEdge relies on docker container runtime. If `docker` is not installed yet, users can install the latest version of docker (for Linux-like systems) with the following command:

```shell
curl -sSL https://get.docker.com | sh
```

After the installation is complete, users can check the version of `docker`:

```shell
docker version
```

**NOTE**：According to the [Official Release Log](https://docs.docker.com/engine/release-notes/#18092), the version of docker lower than 18.09.2 has some security implications. It is recommended to install/update the docker to 18.09.2 and above.

**For more details, please see the [official documentation](https://docs.docker.com/install/).**

## Install OpenEdge

在 OpenEdge 发布新版本的同时，也会发布对应的 rpm、deb 包。使用以下命令，用户可以通过包管理方式将 OpenEdge 的最新版本安装到设备上:

When OpenEdge releases a new version, the rpm and deb packages will be released accordingly. With the following command, users can install the latest version of OpenEdge to the device through package manager:

```shell
curl -sSL http://download.openedge.tech/install.sh | sudo -E bash -
```

After the execution is complete, OpenEdge will be installed to the `/usr/local` directory.

The latest version of OpenEdge uses the Systemd daemon, and users can launch OpenEdge with the following command:

```shell
sudo systemctl start openedge
```

## Import the default configuration package (optional)

As an edge computing platform, OpenEdge provides basic functional modules, such as hub and function-manager modules, in addition to the underlying service management capabilities. Users can edit the configuration files to make `openedge` main program to load corresponding modules and set the running parameters of the modules themselves. For an introduction to each module, refer to [Configuration Interpretation](../tutorials/Config-interpretation.md) for further information.

OpenEdge officially provides a default configuration, users can import the default configuration file by the following command:

```shell
curl -sSL http://download.openedge.tech/install_with_docker_example.sh | sudo -E bash -
```

The default configuration is for learning and testing purposes only. Users should perform on-demand configuration according to actual working scenarios. For details, refer to [Configuration Interpretation](../tutorials/Config-interpretation.md) for further understanding.

If users doesn't need to launch any modules, then there is no need to import any configuration files.

## Verify successful installation

通过包管理器的方式安装 OpenEdge 以后，用户可以依据以下步骤验证 OpenEdge 是否启动成功：

After installing through package manager, users can verify that OpenEdge is successfully installed by following these steps:

- executing the command `ps -ef | grep openedge` to check whether `openedge` is running, as shown below. Otherwise, `openedge` fails to start.

![OpenEdge](../../images/setup/openedge-systemctl-status.png)

- executing the command `docker stats` to view the running status of docker containers. Since the main program `openedge` will first pull the required image to the mirror repository, the user needs to wait 2~5 minutes to execute this command. Take the default configurations in above step for example, the running status of containers are as shown following after pulled. If some containers are missing, it says they failed to start.

![docker stats](../../images/setup/docker-stats.png)

- Under the condition of two above failures, you need to view the log to check more of main program. Log files are stored by default in the `/usr/local/var/log/openedge/openedge.log` director. For errors in the log, the user can refer to the FAQ [FAQ](../FAQ.md). If necessary, just [Submit an issue](https://github.com/baidu/openedge/issues) for help.
