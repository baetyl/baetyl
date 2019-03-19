本文主要提供 OpenEdge 在各系统、平台部署、启动的相关问题及解决方案。

**问题 1**: 在以容器模式启动OpenEdge时，提示缺少启动依赖配置项。

![图片](../images/setup/docker-engine-conf-miss.png)

**参考方案**: 如上图所示，OpenEdge启动缺少配置依赖文件，参考[OpenEdge设计文档](./overview/OpenEdge-design.md)及[GitHub项目开源包](https://github.com/baidu/openedge) example 文件夹补充相应配置文件即可。

**问题 2**: Ubuntu/Debian 下输入命令 `docker info` 后显示如下信息：

```
WARNING: No swap limit support
```

**参考方案**:

1. 修改 `/etc/default/grub` 文件，在其中，编辑或者添加 `GRUB_CMDLINE_LINUX` 为如下内容：
	
	> GRUB_CMDLINE_LINUX="cgroup_enable=memory swapaccount=1"

2. 保存后执行命令 `sudo update-grub`，完成后重启系统生效。

**注意**：如果执行第二步时提示出错，可能是 `grub` 设置有误，请检查后重复步骤1和步骤2.

**问题 3**: WARNING: Your kernel does not support swap limit capabilities. Limitation discarded.

**参考方案**: 参考问题2。

**问题 4**: Got permission denied while trying to connect to the Docker daemon socket at unix:///var/run/docker.sock: Get http://%2Fvar%2Frun%2Fdocker.sock/v1.38/images/json: dial unix /var/run/docker.sock: connect: permission denied.

**参考方案**：

1. 提供管理员权限
2. 通过以下命令添加当前用户到docker用户组

```shell
sudo usermod -aG docker ${USER}
su - ${USER}
``` 

如提示没有 `docker group`，使用如下命令创建新 docker 用户组后再执行上述命令：

```shell
sudo groupadd docker
```

**问题 5**: Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?

**参考方案**：按照问题 4 解决方案执行后如仍报出此问题，重新启动 docker 服务即可。

例，CentOs 下启动命令：

```shell
systemctl start docker
```

**问题 6**: failed to create master: Error response from daemon: client version 1.39 is too new. Maximum supported API version is 1.38.

**参考方案**：设置环境变量 `DOCKER_API_VERSION=1.38`

例，CentOs 下启动命令：

```shell
export DOCKER_API_VERSION=1.38
```

**问题 7**: OpenEdge 如何使用 NB-IoT 连接百度云管理套件或者物接入?

**参考方案**：NB-IoT是一种网络制式，和2/3/4G类似，带宽窄功耗低。NB-IoT支持基于TCP的MQTT通信协议，因此可以使用NB-IoT卡连接百度云物接入，部署OpenEdge应用和BIE云管理通信。但国内三大运营商中，电信对他们的NB卡做了限制，仅允许连接电信的云服务 IP，所以目前只能使用移动NB卡和联通NB卡连接百度云服务。

**问题8**: var/run/openedge.sock: address already in use

**参考方案**：删除 `var/run/openedge.sock`，重启 OpenEdge。

**问题9**: OpenEdge支持数据计算后将计算结果推送Kafka吗？

**参考方案**：支持，您可以[利用Python运行时编写Python脚本](https://github.com/baidu/openedge/blob/master/doc/zh-cn/customize/How-to-write-a-python-script-for-python-runtime.md)，向Hub订阅消息，并将消息逐个写入Kafka。您也可以[自定义模块](https://github.com/baidu/openedge/blob/master/doc/zh-cn/customize/How-to-develop-a-customize-module-for-openedge.md)，该模块向Hub订阅消息，然后批量写入Kafka。
      
**问题10**: Openedge配置更改的方式有哪些？只能通过[云端管理套件](https://cloud.baidu.com/product/bie.html)进行配置更改吗？

目前，我们推荐通过云端管理套件进行配置定义和下发，但您也可以手动更改核心设备上的配置文件，然后重启Openedge使之生效。

**问题 11**：在购买的 NXP LS1046 ARDB 的盒子上部署了 OpenEdge（容器模式），但是启动时候出现 `{"errorDetail": {"message":"no matching manifest for linux/arm64 in the manifest list entries"}, "error":"no matching manifest for linux/arm64 in the manifest list entries"}` 错误信息。

**参考方案**：出现上述问题是由于 OpenEdge 加载执行时候，会根据系统 CPU 类型拉取对应平台镜像启动，而目前 OpenEdge 容器模式暂未支持 Linux/arm64 平台镜像，后续发布版本会加入支持。

**问题 12**：在应用 OpenEdge Hub 模块测试 MQTT 客户端连接时，怎样使用正确的用户名、密码（Hub 模块配置文件存储密码为原明文密码的 SHA256 值）进行连接？

**参考方案**：提供两种方案：（1）在云端管理 Console 创建边缘核心时，显示的窗口中会展示连接使用的用户名和密码（明文，下发后按照 SHA256 值存储），因此，在创建核心时记录即可（**推荐**）；（2）如果启动时候也应用了其他模块，如 Remote 远程模块，Function 函数计算模块，其相应的配置文件中会存储与 Hub 模块连接时应用的用户名和密码（其他模块与 Hub 模块连接，作为 Hub 的客户端），直接获取即可。此外，OpenEdge v0.1.2 将移除该项功能。

**问题 13**：我下载了 Linux MQTTBOX 客户端，解压缩后将可执行文件放置到了 `/usr/local/bin` 目录（其他系统启动加载目录相同，如 `/usr/bin`，`/bin`，`/usr/sbin`等），启动时候提示 `error while loading shared libraries: libgconf-2.so.4: cannot open shared object file: No such file or directory`。

**参考方案**：这是由于 MQTTBOX 启动缺少 `libgconf-2.so.4` 库所致。推荐做法如下：

`步骤 1`: 下载并解压缩 MQTTBOX 软件包；
`步骤 2`: 进入 MQTTBOX 软件包解压缩后的目录，为 MQTTBox 可执行文件配置执行权限；
`步骤 3`：为 MQTTBox 设置软连接：`sudo ln -s /path/to/MQTTBox /usr/local/bin/MQTTBox`；
`步骤 4`：进入终端，执行 `MQTTBox` 即可。

**问题 14**： localfunc 无法进行消息处理，查看 `funclog` 有如下报错信息：

> level=error msg="failed to create new client" dispatcher=mqtt error="dial tcp 0.0.0.0:1883:connect:connection refused"

**参考方案**： 如果是使用BIE云端管理套件下发配置，有如下几个点需要注意：

1. 云端下发配置目前只支持容器模式
2. 如果是云端下发配置，`localfunc` 里配置的hub地址应为 `localhub` 而非 `0.0.0.0`

根据以上信息结合实际报错进行判断，根据需要重新从云端进行配置下发，或者参考[配置解析文档](./tutorials/Config-interpretation.md)进行核对及配置。

**问题 15**： 本地函数计算模块不论发送什么消息，`t/hi` 收到的消息内容都为 `hello world`

**参考方案**： 请查看CFC中Python函数的代码，确定是否有误/Hard Code。