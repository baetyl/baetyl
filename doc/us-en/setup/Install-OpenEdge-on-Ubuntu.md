# Install OpenEdge on Ubuntu

OpenEdge is mainly developed in Go programming language and supports two startup modes: **docker** container mode and **native** process mode.

This document focuses on the installation of the environment required for OpenEdge and the rapid deployment of OpenEdge on the Linux-like system.

**Statement**

- The test system for this article is based on `Ubuntu16.04 amd64`. The kernel and CPU architecture information are viewed by executing the `uname -ar` command as follows:
![ubuntu kernel detail](../../images/setup/os-ubuntu.png)
- In the OpenEdge deployment section, the deployment process is demonstrated using the **docker** container mode.

## Environment Configuration

OpenEdge provides **docker** container mode and **native** process mode. If you are running in **docker** container mode, you need to install the Docker environment; if you are running in **native** process mode, you need to install Python and its runtime dependencies.

### Install Docker in **docker** container mode

To start using **docker** container mode (recommended), you need to complete the docker installation first.

**NOTE**:

- The official Dockerfile is offered for multi-stage builds. If you need to build the relevant image yourself, The version of `Docker` you installed should be above 17.05.
- The production environment can run the image using a lower version of `Docker`, which is currently tested to a minimum usable version of 12.0.
- According to the [Official Release Log](https://docs.docker.com/engine/release-notes/#18092), the version of `Docker` lower than 18.09.2 has some security implications. It is recommended to install/update `Docker` to 18.09.2 and above.

Can be installed by the following command(Suitable for linux-like systems, [Supported Platforms](./Support-platforms.md)):

```shell
curl -sSL https://get.docker.com | sh
```

Or install by using following command:

```shell
sudo snap install docker # Ubuntu16.04 after
sudo apt-get install docker.io # Ubuntu 16.04 before
```

**NOTE**:

- After the `Docker` installation is complete, use the following command to view the installed version of `Docker`.

```shell
docker version
```

**For more details, please see the [official documentation](https://docs.docker.com/install/).**

### Install Python and runtime dependency package in **native** process mode

OpenEdge provides Python Runtime, which supports running code written in Python2.7 and Python3.6. If you plan to use the **native** process mode to start, it is recommended to install **Python3.6** or higher version locally and run the package it depends on. If you already have other version of Python3 lower than 3.6, it is recommended that you uninstall it first and install Python3.6. Or you can keep the inconsistent version but need to ensure compatibility.

- Step 1：Check the version of Python3. If the version is Python3.6 or higher, jump Step 3. Instead, jump Step 2.

```shell
which python3
```

- Step 2：Use the following commands to install Python3.6.

```shell
sudo add-apt-repository ppa:jonathonf/python-3.6
sudo apt-get update
sudo apt-get install python3.6
sudo apt-get install python3-pip
sudo pip3 install pyyaml protobuf grpcio
```

- Step 3：Install the packages need by OpenEdge based on Python3.6.

```shell
sudo pip3 install grpcio protobuf pyyaml
```

## Deploy OpenEdge

### Deployment Process

- Step1: Download [OpenEdge](../Resources-download.md);
- Step2: Open the terminal and enter the OpenEdge directory for decompression:

```shell
tar -zxvf openedge-xxx.tar.gz
```

- Step3: After the decompression operation is completed, execute the command `sudo openedge start` in the OpenEdge directory to start OpenEdge. Then check the starting and loading logs, meantimes execute the command `docker stats` to display the running status of the docker containers. Compare both to see whether all the images needed by OpenEdge are loaded successfully by docker.
- Step4: If the images to be launched in logs are all successfully loaded by the docker containers, OpenEdge is successfully started.

**NOTE**: The official download page only provides the docker mode executable file. If you want to run in process mode, please refer to [Build-OpenEdge-From-Source](./Build-OpenEdge-from-Source.md)

### Start Deployment

As mentioned above, download OpenEdge from the [Download page](../Resources-download.md) first (also can compile from source, see [Build-OpenEdge-From-Source](./Build-OpenEdge-from-Source.md)), then open the terminal to enter OpenEdge directory for decompression. After successful decompression, you can find that the openedge directory mainly includes `bin`, `etc`, `var`, etc., as shown in the following picture:

![OpenEdge directory](../../images/setup/openedge-dir-ubuntu.png)

The `bin` directory stores the openedge executable binary file, the `etc` directory stores the configuration of OpenEdge, and the `var` directory stores the configuration and resources for the modules of OpenEdge.

Place the binary file under `/usr/local/bin` or any directory that exists in your environment variable's `PATH` value. And copy the `etc` and `var` directories to the `/usr/local` or other upper level directories where you place the executable. Of course, you can just leave them in the place where you unpacked.

Then, open a new terminal and execute the command `docker stats` to view the running status of the container in the installed docker, as shown in the following picture:

![view the docker containers status](../../images/setup/docker-stats-before-ubuntu.png)

It can be found that the current system does not have a docker container running.

Then, step into the decompressed folder of OpenEdge, execute the command `sudo openedge start` and if you didn't put the `var` and `etc` directories to the upper level directory of where you keep executable file, you need use `-w` to specify the work directory like this: `sudo openedge start -w yourpath/to/configuration` . Check the result by executing the command `ps -ef | grep "openedge"` , as shown below:

![OpenEdge startup log](../../images/setup/openedge-started-thread-ubuntu.png)

And you can check the log file for details. Log files are stored by default in the `var/log/openedge` directory of the working directory.

At the same time, observe the terminal that shows the running status of the container, as shown in the following picture:

![running containers](../../images/setup/docker-stats-after-ubuntu.png)

Obviously, OpenEdge has been successfully launched.

As mentioned above, if the steps are executed correctly, OpenEdge can be quickly deployed and started on the Ubuntu system.