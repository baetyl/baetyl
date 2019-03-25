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

_**NOTE**: If you got some error when you execute `step2`, it may be that the `grub` setting is incorrect. Please repeat `steps 1 and 2`._

**Question 3**: WARNING: Your kernel does not support swap limit capabilities. Limitation discarded.

**Suggested Solution**: Refer to Question 2.

**Question 4**: Got permission denied while trying to connect to the Docker daemon socket at unix:///var/run/docker.sock: Get http://%2Fvar%2Frun%2Fdocker.sock/v1.38/images/json: dial unix /var/run/docker.sock: connect: permission denied.

**Suggested Solution**: Add the docker group if it doesn't already exist:

```shell
sudo groupadd docker
```

Add the current user to the docker group:

```shell
sudo usermod -aG docker ${USER}
su - ${USER}
``` 

**Question 5**: Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?

**Suggested Solution**: If you still report this issue after the solution of Question 4 solution is executed, restart the docker service.

For example, execute the following command on CentOs:

```shell
systemctl start docker
```

**Question 6**: failed to create master: Error response from daemon: client version 1.39 is too new. Maximum supported API version is 1.38.

**Suggested Solution**: Workaround is to pass API version via environment variable:

`DOCKER_API_VERSION=1.38`

For example, execute the following command on CentOs:

```shell
export DOCKER_API_VERSION=1.38
```

**Question 7**: How does OpenEdge connect to NB-IOT network?

**Suggested Solution**: NB-IoT is a network standard similar to 2/3/4G with low bandwidth and low power consumption. NB-IoT supports TCP-based MQTT protocol, so you can use NB-IoT card to connect to Baidu Cloud IotHub, deploy OpenEdge application and communicate with [BIE](https://cloud.baidu.com/product/bie.html) Cloud Management Suite. However, among the three major operators in China, Telecom have imposed whitelist restrictions on their NB cards, and only allow to connect to Telecom Cloud Service IP. Therefore, only Mobile NB cards and Unicom NB cards can be used to connect to Baidu Cloud Service.

**Question 8**: var/run/openedge.sock: address already in use

**Suggested Solution**: Remove var/run/openedge.sock and restart OpenEdge.

**Question 9**: Does OpenEdge support to push data to Kafka?

**Suggested Solution**: For support, you can use the local function module to write a [Python script](https://github.com/baidu/openedge/blob/master/doc/us-en/customize/How-to-write-a-python-script-for-python-runtime.md) that is responsible for subscribing messages from the local Hub module and writing them to Kafka service. Besides, you can also develop a [customize module](https://github.com/baidu/openedge/blob/master/doc/us-en/customize/How-to-develop-a-customize-module-for-openedge.md), which subscribes message from the local Hub module and then writes it to Kafka.

**Question 10**: What are the ways to change OpenEdge configurations? Can I only make configuration changes through the [BIE](https://cloud.baidu.com/product/bie.html) Cloud Management Suite?

**Suggested Solution**: Currently, we recommend changing configurations through the BIE Cloud Management Suite, but you can also manually change the configuration file on the core device and then restart OpenEdge to take effect.

**Question 11**：I deploy OpenEdge on NXP LS1046 ARDB box，but it reports an error of `{"errorDetail": {"message":"no matching manifest for linux/arm64 in the manifest list entries"}, "error":"no matching manifest for linux/arm64 in the manifest list entries"}` when OpenEdge start.

**Suggested Solution**：The above problem occurs because the OpenEdge start will pull the module image due to manifest(the system CPU type). And now, OpenEdge does not support the Linux/arm64 docker image, and subsequent releases will be supported.

**Question 12**：When using local Hub module to test an MQTT client's connection, how do I get the correct username and password (the Hub module configuration file stores the password as its SHA256 value)?

**Suggested Solution**：Two solutions are provided: (1) When the Edge Management Core is created in the [Cloud Management Console](https://cloud.baidu.com/product/bie.html), the connected username and password are displayed in the window (plain text, stored its(password) SHA256 value when deploy), so you can record when creating the core(**Recommended**); (2) If other modules are also applied at startup, such as the Remote module, the Function module, the corresponding configuration file stores the username and password applied when connecting to the Hub module (other module as MQTT client when connected to Hub module), and it can be obtained directly. Besides, OpenEdge v0.1.2 will remove it.

**Question 13**：I download MQTTBOX client, extract it to a directory, and copy/move the executable file `MQTTBox` to `/usr/local/bin`(other directory is similar, such as `/usr/bin`, `/bin`, `/usr/sbin`, etc.). But it reports an error of `error while loading shared libraries: libgconf-2.so.4: cannot open shared object file: No such file or directory` when `MQTTBox` start.

**Suggested Solution**：As above description, this is because the lack of `libgconf-2.so.4` library when `MQTTBox` start, and the recommended use is as follows:

`Step 1`: Download and extract the MQTTBOX software package;
`Step 2`: `cd /pat/to/MQTTBOX/directory and sudo chmod +x MQTTBox`;
`Step 3`：`sudo ln -s /path/to/MQTTBox /usr/local/bin/MQTTBox`;
`Step 4`：Open terminal and execute the command `MQTTBox`.

**Question 14**: localfunc can't process the message, check `funclog` has the following error message:

> level=error msg="failed to create new client" dispatcher=mqtt error="dial tcp 0.0.0.0:1883:connect:connection refused"

**Suggested Solution**: If you are using the BIE Cloud Management Suite to deliver the configuration, there are a few points to note:

1. Cloud delivery configuration currently only supports container mode.
2. If the configuration is sent in the cloud, the hub address configured in `localfunc` should be `localhub` instead of `0.0.0.0`.

According to the above information, the actual error is judged, and the configuration is delivered from the cloud as needed, or by referring to [Configuration Analysis Document](./tutorials/Config-interpretation.md) for verification and configuration.

**Question 15**: The local function calculation module receives the message, `t/hi` receives the message content as `hello world`.

**Suggested Solution**: Please check the code of the Python function in CFC to determine if there is a mistake/Hard Code.

**Question 16**： How can i use BIE Cloud Management Suite with [CFC(Cloud Function Compute)](https://cloud.baidu.com/product/cfc.html)?

**Suggested Solution**： 
1. Make sure your BIE configuration and CFC functions in the same region, such as beijing/guangzhou.
2. Make sure your CFC functions are published.

**Question 17**： What‘s the relationship between the parameter expose  and the parameter listen which in the hub configuration file?

**Answer**： 
1. expose: Port exposed configuration in Docker container mode.
2. listen: Which address the hub module will listen on. In docker container mode, it's means container address. In native process mode, it's means host address.
3. By referring to [Configuration Analysis Document](./tutorials/Config-interpretation.md)