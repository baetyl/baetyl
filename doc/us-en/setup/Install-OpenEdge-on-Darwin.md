# Install OpenEdge On Darwin

> OpenEdge is mainly developed in Go programming language and supports two startup modes: ***docker*** container mode and ***native*** process mode.

This document focuses on the installation and configuration of the environment required for OpenEdge and the rapid deployment of OpenEdge on the Darwin system.

## Environment Configuration

### Install Docker

> OpenEdge offers two startup modes. To start using ***docker*** container mode (recommended), you need to complete the docker installation first.

**Notice:**

+ The official Dockerfile is offered for multi-stage builds. If you need to build the relevant image yourself, The version of docker you installed should be above 17.05.
+ The production environment can run the image using a lower version of Docker, which is currently tested to a minimum usable version of 12.0.
+ According to the [Official Release Log](https://docs.docker.com/engine/release-notes/#18092), the version of docker version < 18.09.2 has some security implications. It is recommended to install/update the Docker version to 18.09. 2 and above.

Go to [official page](https://hub.docker.com/editions/community/docker-ce-desktop-mac) to download the .dmg file you need. Once done, double-click to open and drag Docker into the Application folder.

![Install On Darwin](../../images/setup/docker-install-on-mac.png)

View the version of installed Docker:

```shell
docker version
```

**For more details, please see the [official documentation](https://docs.docker.com/install/).**

### Install Python2.7 and Python runtime dependency package

> + OpenEdge provides Python Runtime, which supports running code written in Python version 2.7. If you plan to start using ***native*** process mode, you need to install Python 2.7 and the package you depend on. If you plan to start in ***docker*** container mode, you do not need to perform the following steps.

Install by using HomeBrew(Recommened):

```shell
/usr/bin/ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"  // Install HomeBrew
brew install python@2
pip isntall protobuf grpcio
```

***Notice***: Execute the following command to check the installed Python version:

```shell
python -V
```
 
Specify the default Python version with following command:

```shell
alias python=/yourpath/python2.7
```

## Deploy OpenEdge

### Preparation Before Deployment

**Statement**:

+ The following is an example of the deployment and startup of OpenEdge on Darwin system. It is assumed that the environment required for OpenEdge operation has been configured [before](#Environment-Configuration).
+ The Darwin system mentioned below is based on the Darwin High Sierra Version 10.13.6 with a CPU architecture of x86_64. Execute command `uname -ar` and get the result like this:

![darwin kernel detail](../../images/setup/os-darwin.png)

Starting OpenEdge containerization mode requires the running device to complete the installation and operation of Docker. You can install it by referring to [Steps above](#Install-Docker).

### Deployment Process

- **Step1**: [Download](../Resources-download.md) OpenEdge archive;
- **Step2**: Open the terminal and enter the OpenEdge directory for decompression:
	- execute command `tar -zxvf openedge-xxx.tar.gz`;
- **Step3**: After the decompression operation is completed, enter the OpenEdge package directory in the terminal, open a new terminal at the same time, execute the command `docker stats`, display the running status of the container in the installted docker, and then execute the command `bin/openedge -w .`, respectively. Observe the contents displayed by the two terminals;
- **Step4**: If the results are consistent, it means that OpenEdge has started normally.

***Notice:*** The official download page only provides the docker mode executable file. If you want to run in process mode, please refer to [Build-OpenEdge-From-Source.md](./Build-OpenEdge-from-Source.md)

### Start deployment

As mentioned above, download OpenEdge archive from the [Download page](../Resources-download.md) first (also can compile from source, see [Build-OpenEdge-From-Source.md](./Build-OpenEdge-from-Source.md)), then open the terminal to enter OpenEdge directory for decompression. After successful decompression, you can find that the openedge directory mainly includes bin, etc, var, etc., as shown in the following picture:

![OpenEdge directory](../../images/setup/openedge-dir.png)

The bin directory stores the openedge executable binary file, the etc directory stores the configuration of OpenEdge, and the var directory stores the configuration and resources for the modules of OpenEdge.

Then, open a new terminal and execute the command `docker stats` to view the running status of the container in the installted docker, as shown in the following picture:

![view the docker containers status](../../images/setup/docker-stats-before.png)

It can be found that the current system does not have a docker container running.

Then, go to the decompressed OpenEdge folder, execute the command `bin/openedge -w .` in the another terminal, observe the log of OpenEdge startup, as shown below:

![OpenEdge startup log](../../images/setup/docker-openedge-start.png)

At the same time, observe the terminal that shows the running status of the container, as shown in the following picture:

![running containers](../../images/setup/docker-stats-after.png)

Obviously, OpenEdge has been successfully launched.

As mentioned above, if the steps are executed correctly, OpenEdge can be quickly deployed and started on the Darwin system.