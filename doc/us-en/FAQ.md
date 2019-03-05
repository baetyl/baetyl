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

**Question 7**: How does BIE access the NB-IOT network?

NB-IoT is a network standard similar to 2/3/4G with low bandwidth and low power consumption. NB-IoT supports TCP-based MQTT protocol, so you can use NB-IoT card to connect to Baidu Cloud IotHub, deploy OpenEdge application and communicate with BIE Cloud Management Suite. However, among the three major operators in China, Telecom have imposed whitelist restrictions on their NB cards, and only allow to connect to Telecom Cloud Service IP. Therefore, only Mobile NB cards and Unicom NB cards can be used to connect to Baidu Cloud Service.

**Question 8**: var/run/openedge.sock: address already in use

Remove var/run/openedge.sock and restart Openedge.

**Question 9**: Does openedge support to push data to kafka?

For support, you can use the function framework to write a function that is responsible for subscribing messages from the Hub and writing them to Kafka one by one. You can also customize the module, which subscribes to the Hub and then writes kafka in bulk.

**Question 10**: What are the ways to change openedge configurations? Can I only make configuration changes through the BIE Cloud Management Suite?

Currently, we recommend changing configurations through the BIE Cloud Management Suite, but you can also manually change the configuration file on the core device and then restart openedge to take effect.

**Question 11**: Does the openedge function framework support multiple running instances of function? How to configure?

The function framework would start multiple running instances based on the real-time computing load. The core configuration is as follows:
```
instance: function instance configuration
  min: The default value is `0`, means the minimum number of function instance. And the minimum configuration allowed to be set is `0`, the maximum configuration allowed to be set is `100`.
  max: The default value is `1`, means the maximum number of function instance. And the minimum configuration allowed to be set is `1`, the maximum configuration allowed to be set is `100`.
```
For more function configuration: [Local Function Configuration](https://github.com/baidu/openedge/blob/master/doc/us-en/tutorials/Config-interpretation.md#local-function-configuration)