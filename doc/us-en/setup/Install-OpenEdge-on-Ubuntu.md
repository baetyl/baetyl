# Install OpenEdge on Ubuntu

> OpenEdge is mainly developed in Go programming language and supports two startup modes: ***docker*** container mode and ***native*** process mode.

This document focuses on the installation and configuration of the environment required for OpenEdge and the rapid deployment of OpenEdge on the Linux-like system.

## Environment Configuration

### Install Docker

> OpenEdge offers two startup modes. To start using ***docker*** container mode (recommended), you need to complete the docker installation first.

**Notice:**

+ The official Dockerfile is offered for multi-stage builds. If you need to build the relevant image yourself, The version of docker you installed should be above 17.05.
+ The production environment can run the image using a lower version of Docker, which is currently tested to a minimum usable version of 12.0.
+ According to the [Official Release Log](https://docs.docker.com/engine/release-notes/#18092), the version of Docker lower than 18.09.2 has some security implications. It is recommended to install/update the Docker to 18.09.2 and above.

Can be installed by the following command(Suitable for linux-like systems, [Supported Platforms](./Support-platforms.md)):

```shell
curl -sSL https://get.docker.com | sh
```

Or install by using following command:

```shell
sudo snap install docker # Ubuntu16.04 after
sudo apt-get install docker.io # Ubuntu 16.04 before
```

***Notice***: 

+ After the Docker installation is complete, use the following command to view the installed version of Docker.

```shell
docker version
```

**For more details, please see the [official documentation](https://docs.docker.com/install/).**

### Install Python2.7 and Python runtime dependency package

> + OpenEdge provides Python Runtime, which supports running code written in Python2.7. If you run OpenEdge in **native process mode**, you **MUST** firstly install Python2.7 and the package actually use. But, If you plan to start in ***docker container mode***, you do not need to perform the following steps.

Commands:

```shell
sudo apt-get update
sudo apt-get upgrade
sudo apt-get install python2.7
sudo apt-get install python-pip
sudo pip install protobuf grpcio
```

> Execute the command `python -V` to see that the version of Python is 2.7.* and the installation is correct.

### Specify The Default Version Of Python

In some cases, you need to specify the default version of Python for the above installed version. Complete with the following command (Valid after reboot):

```shell
alias python=/yourpath/python2.7
```

## Deploy OpenEdge

### Preparation Before Deployment

**Statement**:

+ The following is an example of the deployment and startup of OpenEdge on Ubuntu system. It is assumed that the environment required for OpenEdge operation has been [configured](#Environment-Configuration).
+ The Ubuntu system mentioned below is based on the version of kernel is 4.15.0-29-generic, and the CPU architecture is x86_64. Then execute the command `uname -ar` and get the result like this:

![ubuntu kernel detail](../../images/setup/os-ubuntu.png)

Starting OpenEdge containerization mode requires the running device to complete the installation and operation of Docker. You can install it by referring to [Steps above](#Install-Docker).

### Deployment Process

- **Step1**: [Download](../Resources-download.md) OpenEdge;
- **Step2**: Open the terminal and enter the OpenEdge directory for decompression:
	- Execute the command `tar -zxvf openedge-xxx.tar.gz`;
- **Step3**: After the decompression operation is completed, enter the OpenEdge package directory in the terminal, open a new terminal at the same time, execute the command `docker stats`, display the running status of the container in the installed docker, and then execute the command `sudo ./bin/openedge start`, respectively. Observe the contents displayed by the two terminals;
- **Step4**: If the results are consistent, it means that OpenEdge has started normally.

***Notice:*** The official download page only provides the docker mode executable file. If you want to run in process mode, please refer to [Build-OpenEdge-From-Source.md](./Build-OpenEdge-from-Source.md)

### Start Deployment

As mentioned above, download OpenEdge from the [Download page](../Resources-download.md) first (also can compile from source, see [Build-OpenEdge-From-Source.md](./Build-OpenEdge-from-Source.md)), then open the terminal to enter OpenEdge directory for decompression. After successful decompression, you can find that the openedge directory mainly includes `bin`, `etc`, `var`, etc., as shown in the following picture:

![OpenEdge directory](../../images/setup/openedge-dir-ubuntu.png)

The `bin` directory stores the openedge executable binary file, the `etc` directory stores the configuration of OpenEdge, and the `var` directory stores the configuration and resources for the modules of OpenEdge.

Then, open a new terminal and execute the command `docker stats` to view the running status of the container in the installed docker, as shown in the following picture:

![view the docker containers status](../../images/setup/docker-stats-before-ubuntu.png)

It can be found that the current system does not have a docker container running.

Then, step into the decompressed folder of OpenEdge, execute the command `sudo ./bin/openedge start` in the another terminal, check the result by executing the command `ps -ef | grep "openedge"` , as shown below:

![OpenEdge startup log](../../images/setup/openedge-started-thread-ubuntu.png)

At the same time, observe the terminal that shows the running status of the container, as shown in the following picture:

![running containers](../../images/setup/docker-stats-after-ubuntu.png)

Obviously, OpenEdge has been successfully launched.

As mentioned above, if the steps are executed correctly, OpenEdge can be quickly deployed and started on the Ubuntu system.