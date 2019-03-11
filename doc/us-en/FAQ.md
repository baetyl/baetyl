This document mainly provides related issues and solutions for OpenEdge deployment and startup in various platforms.

**Question 1**: Prompt missing startup dependency configuration item when starting OpenEdge in docker container mode.

![Picture](../images/setup/docker-engine-conf-miss.png)

**Suggested Solution**: As shown in the above picture, OpenEdge startup lacks configuration dependency files, refer to [OpenEdge Design Document](./overview/OpenEdge-design.md) and [GitHub Project Open Source Package](https://github.com/baidu/openedge) to supplemented with the corresponding configuration file in the folder named `example`.

**Question 2**: Execute the command `docker info` get the following result on Ubuntu/Debian:

```
WARNING: No swap limit support
```

**Suggested Solution**:

1. Open `/etc/default/grub` with your favorite text editor. Make sure the following lines are commented out or add them if they don't exist:

	> GRUB_CMDLINE_LINUX="cgroup_enable=memory swapaccount=1"

2. Save and exit and then run: `sudo update-grub` and reboot.

***Notice:*** If you got some error when you execute step2, it may be that the `grub` setting is incorrect. Please repeat steps 1 and 2.

**Question 3**: WARNING: Your kernel does not support swap limit capabilities. Limitation discarded.

**Suggested Solution**: Refer to Question 2.

**Question 4**: Got permission denied while trying to connect to the Docker daemon socket at unix:///var/run/docker.sock: Get http://%2Fvar%2Frun%2Fdocker.sock/v1.38/images/json: dial unix /var/run/docker.sock: connect: permission denied.

Add the docker group if it doesn't already exist:

```shell
sudo groupadd docker
```

Add the current user to the docker group:

```shell
sudo usermod -aG docker ${USER}
su - ${USER}
``` 

**Question 5**: Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?

If you still report this issue after the solution of Question 4 solution is executed, restart the docker service.

For example, execute the following command on CentOs:

```shell
systemctl start docker
```

**Question 6**: failed to create master: Error response from daemon: client version 1.39 is too new. Maximum supported API version is 1.38.

workaround is to pass API version via environment variable:
DOCKER_API_VERSION=1.38

For example, execute the following command on CentOs:

```shell
export DOCKER_API_VERSION=1.38
```

**Question 7**: How does OpenEdge connect to NB-IOT network?

NB-IoT is a network standard similar to 2/3/4G with low bandwidth and low power consumption. NB-IoT supports TCP-based MQTT protocol, so you can use NB-IoT card to connect to Baidu Cloud IotHub, deploy OpenEdge application and communicate with [BIE](https://cloud.baidu.com/product/bie.html) Cloud Management Suite. However, among the three major operators in China, Telecom have imposed whitelist restrictions on their NB cards, and only allow to connect to Telecom Cloud Service IP. Therefore, only Mobile NB cards and Unicom NB cards can be used to connect to Baidu Cloud Service.

**Question 8**: var/run/openedge.sock: address already in use

Remove var/run/openedge.sock and restart OpenEdge.

**Question 9**: Does OpenEdge support to push data to Kafka?

For support, you can use the local function module to write a [Python script](https://github.com/baidu/openedge/blob/master/doc/us-en/customize/How-to-write-a-python-script-for-python-runtime.md) that is responsible for subscribing messages from the local Hub module and writing them to Kafka service. Besides, you can also develop a [customize module](https://github.com/baidu/openedge/blob/master/doc/us-en/customize/How-to-develop-a-customize-module-for-openedge.md), which subscribes message from the local Hub module and then writes it to Kafka.

**Question 10**: What are the ways to change OpenEdge configurations? Can I only make configuration changes through the [BIE](https://cloud.baidu.com/product/bie.html) Cloud Management Suite?

Currently, we recommend changing configurations through the BIE Cloud Management Suite, but you can also manually change the configuration file on the core device and then restart OpenEdge to take effect.
