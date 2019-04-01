# OpenEdge

- [基本概念](#基本概念)
- [系统组成](#系统组成)
- [主程序](#主程序)
  - [引擎系统](#引擎系统)
    - [Docker引擎](#docker引擎)
    - [Native引擎](#native引擎)
  - [RESTful API](#restful-api)
    - [/system/inspect](#systeminspect)
    - [/system/update](#systemupdate)
    - [/services/{}/instances/{}/start|stop](#servicesinstancesstartstop)
  - [环境变量](#环境变量)
- [官方模块](#官方模块)
  - [openedge-agent](#openedge-agent)
  - [openedge-hub](#openedge-hub)
  - [openedge-function-manager](#openedge-function-manager)
  - [openedge-function-python27](#openedge-function-python27)
  - [openedge-remote-mqtt](#openedge-remote-mqtt)

## 基本概念

- **系统**：这里专指OpenEdge系统，包行**主程序**、**服务**、**存储卷**和使用的系统资源。
- **主程序**： 指OpenEdge实现的核心部分，负责管理所有**存储卷**和**服务**，内置**引擎系统**，对外提供RESTful API和命令行等。
- **服务**：指一组接受OpenEdge控制的运行程序集合，用于提供某些具体的功能，比如消息路由服务、函数计算服务、微服务等。
- **实例**：指**服务**启动的具体的运行程序或容器，一个**服务**可以启动多个实例，也可以不启动实例由其他服务负责动态启动实例，比如函数计算的运行时实例就是由函数计算管理服务动态启停的。
- **存储卷**：指被**服务**使用的目录，可以是只读的目录，比如放置配置、证书、脚本等资源的目录，也可以是可写的目录，比如日志、数据等持久化目录。
- **引擎系统**： 指**服务**的各类运行模式的操作抽象和具体实现，比如docker容器模式和native进程模式。
- **服务**和**系统**的关系：OpenEdge系统可以启动多个服务，服务之间没有依赖关系，不应当假设他们的启动顺序（虽然当前还是顺序启动的）。服务在运行时产生的所有信息都是临时的，除非映射到持久化目录，服务停止后这些信息都会被删除。服务内的程序由于种种原因可能会停止，服务会根据用户的配置对程序进行重启，这种情况不等于服务的停止，所以信息不会被删除。

## 系统组成

一个完整的OpenEdge系统由**主程序**、**服务**、**存储卷**和使用的系统资源组成，主程序根据应用配置加载各个模块启动相应的服务，一个服务又可以启动若干实例，所有实例都由主程序负责管理和守护。需要注意的是同一个服务下的实例共享所有存储卷，所以如果出现独占的资源，比如监听同一个端口，使用同一个Client ID连接Hub等，只能启动一个实例。

目前OpenEdge开源了如下几个官方模块：

- [openedge-agent](#openedge-agent)：提供BIE云代理服务，进行状态上报和应用下发。
- [openedge-hub](#openedge-hub)：提供基于MQTT的消息路由服务。
- [openedge-remote-mqtt](#openedge-remote-mqtt)：提供同远程MQTT服务进行消息同步的服务。
- [openedge-function-manager](#openedge-function-manager)：提供函数计算服务，进行函数实例管理和消息触发的函数调用。
- [openedge-function-python27](#openedge-function-python27)：提供加载Python脚本的GRPC微服务，可以托管给 openedge-function-manager 成为函数实例提供方。

结构图:

![结构图](../../images/overview/design/openedge_design.png)

## 主程序

**主程序**作为OpenEdge系统的核心，负责管理所有存储卷和服务，内置运行引擎系统，对外提供RESTful API和命令行。

主程序启停过程如下：

1. 执行启动命令：`sudo openedge start`，默认工作目录为openedge安装目录的上一级目录。
2. 主程序首先会加载工作目录下的etc/openedge/openedge.yml，初始化运行模式、API Server、日志和退出超时时间等，这些配置不会随应用配置下发而改变。如果启动没有报错，会在/var/run/目录下生成openedge.pid和openedge.sock（Linux）文件。
3. 然后主程序会尝试加载应用配置var/db/openedge/application.yml，如果该配置不存在则不启动任何服务，否则加载应用配置中的服务列表和存储卷列表。该文件会随应用配置下发而更新，届时系统会根据新配置重新编排服务。
4. 在启动所有服务前，主程序会先调用Engine接口执行一些准备工作，比如容器模式下会先尝试下载所有服务的镜像。
5. 准备工作完成后，开始顺序启动所有服务，如果服务启动失败则会导致主程序退出。容器模式下会将存储卷映射到容器内部；进程模式下会为每个服务创建各自的临时工作目录，并将存储卷软链到工作目录下，服务退出后临时工作目录会被清理，行为同容器模式。
6. 最后，如果需要退出，可执行 `sudo openedge stop`，主程序会同时通知所有服务的实例退出并等待，如果超时则强制杀掉实例。然后清理openedge.pid和openedge.sock（Linux）后退出。

下面是完整的application.yml配置字段：

```golang
// 应用配置
type AppConfig struct {
    // 指定应用配置的版本号
    Version  string        `yaml:"version" json:"version"`
    // 指定应用的所以服务信息
    Services []ServiceInfo `yaml:"services" json:"services" default:"[]"`
    // 指定应用的所以存储卷信息
    Volumes  []VolumeInfo  `yaml:"volumes" json:"volumes" default:"[]"`
}

// 存储卷配置
type VolumeInfo struct {
    // 指定存储卷的唯一名称
    Name     string `yaml:"name" json:"name" validate:"regexp=^[a-zA-Z0-9][a-zA-Z0-9_-]{0\\,63}$"`
    // 指定存储卷在宿主机上的目录
    Path     string `yaml:"path" json:"path" validate:"nonzero"`
}

// 存储卷映射配置
type MountInfo struct {
    // 被映射存储卷的名称
    Name     string `yaml:"name" json:"name" validate:"regexp=^[a-zA-Z0-9][a-zA-Z0-9_-]{0\\,63}$"`
    // 指定存储卷在容器内的目录
    Path     string `yaml:"path" json:"path" validate:"nonzero"`
    // 指定存储卷的操作权限，只读或可写
    ReadOnly bool   `yaml:"readonly" json:"readonly"`
}

// 服务配置
type ServiceInfo struct {
    // 指定服务的唯一名称，在整个OpenEdge系统内唯一，不同的OpenEdge系统间不要求唯一
    Name      string            `yaml:"name" json:"name" validate:"regexp=^[a-zA-Z0-9][a-zA-Z0-9_-]{0\\,63}$"`
    // 指定服务的程序地址，通常使用Docker镜像名称
    Image     string            `yaml:"image" json:"image" validate:"nonzero"`
    // 指定服务副本书，即启动的实例数
    Replica   int               `yaml:"replica" json:"replica"  validate:"min=0"`
    // 指定服务需要映射的存储卷，将存储卷映射到容器中目录
    Mounts    []MountInfo       `yaml:"mounts" json:"mounts" default:"[]"`
    // 指定服务对外暴露的端口号，用于Docker容器模式
    Ports     []string          `yaml:"ports" json:"ports" default:"[]"`
    // 指定服务需要映射的设备，用于Docker容器模式
    Devices   []string          `yaml:"devices" json:"devices" default:"[]"`
    // 指定服务程序的启动参数，但不包括`arg[0]`
    Args      []string          `yaml:"args" json:"args" default:"[]"`
    // 指定服务程序的环境变量，但不包含系统的环境变量
    Env       map[string]string `yaml:"env" json:"env" default:"{}"`
    // 指定服务重启策略
    Restart   RestartPolicyInfo `yaml:"restart" json:"restart"`
    // 指定服务单个实例的资源限制，用于Docker容器模式
    Resources Resources         `yaml:"resources" json:"resources"`
}
```

### 引擎系统

**引擎系统**负责服务的存储卷映射，实例启停、监听和守护等，对服务操作做了抽象，可以实现不同的服务运行模式，根据运行模式的不同，系统会使用不同的引擎（Engine）来启动服务。目前支持了docker容器模式和native进程模式，后续还会支持k3s容器模式。

#### Docker引擎

Docker引擎会将服务Image解释为Docker镜像地址，并通过操纵`docker`来启动服务。所有服务的网络都会位于一个Docker Network中，并根据Ports信息来对外暴露端口，根据Mounts信息来映射目录，根据Devices信息来映射设备，根据Resources信息来配置容器可使用的资源，比如CPU、内存等。服务间的访问直接使用服务名称即可，由Docker的DNS Server负责路由。服务的每个实例对应于一个容器，引擎负责容器的启停和重启。

#### Native引擎

在无法提供容器服务的平台（如旧版本的Windows）上，Native引擎以裸进程方式尽可能的模拟容器的使用体验。该引擎会将服务Image解释为Package名称，Package由存储卷提供，内含服务所需的程序，但这些程序的依赖（如Python解释器、lib等）需要在主机上提前安装好。所有服务的网络都是主机网络，所有端口都是暴露的，用户需要注意避免端口冲突。服务的每个实例对应于一个进程，引擎负责进程的启停和重启。

_**注意**：进程模式不支持资源的限制，无需暴露端口、映射设备。_

目前上诉两种模式基本实现了配置统一，唯一一遗留了服务地址配置差异部分，所以example中的配置分成了native和docker两个目录，但最终会实现统一。

### RESTful API

OpenEdge主程序会暴露一组RESTful API，在Linux系统下默认采用Unix Domain Socket，固定地址为`/var/openedge.sock`；其他环境采用TCP，默认地址为`tcp://127.0.0.1:50050`。目前接口的认证方式采用简单的动态Token的方式，主程序在启动服务时会为每个服务动态生成一个Token，将服务名和Token以环境变量的方式传入服务的实例，实例读取后放入请求的Header中发给主程序即可。需要注意的是动态启动的实例是无法获取到Token的，因此动态实例无法动态启动其他实例。

Header名称如下：

- x-iot-edge-username：账号名称，即服务名
- x-iot-edge-password：账号密码，即动态Token

下面是目前提供的接口：

- GET /system/inspect 获取系统信息和状态
- PUT /system/update 更新系统
- GET /ports/available 获取空闲端口
- PUT /services/{serviceName}/instances/{instanceName}/start 动态启动某个服务的一个实例
- PUT /services/{serviceName}/instances/{instanceName}/stop 动态停止某个服务的某个实例

#### /system/inspect

该接口用于获取如下信息和状态：

```golang
// 采集的所有信息
type Inspect struct {
    // 异常信息
    Error    string    `json:"error,omitempty"`
    // 采集时间
    Time     time.Time `json:"time,omitempty"`
    // 软件信息
    Software Software  `json:"software,omitempty"`
    // 硬件消息
    Hardware Hardware  `json:"hardware,omitempty"`
    // 服务信息，包括服务名、实例运行状态等
    Services Services  `json:"services,omitempty"`
    // 存储卷信息，当前未实现
    // Volumes  []VolumeStatus `json:"volumes,omitempty"`
}

// 软件信息
type Software struct {
    // 宿主机系统消息
    OS          string `json:"os,omitempty"`
    // 宿主机CPU型号
    Arch        string `json:"arch,omitempty"`
    // OpenEdge运行模式
    Mode        string `json:"mode,omitempty"`
    // OpenEdge编译的Golang版本
    GoVersion   string `json:"go_version,omitempty"`
    // OpenEdge发布版本
    BinVersion  string `json:"bin_version,omitempty"`
    // OpenEdge运行的应用配置版本
    ConfVersion string `json:"conf_version,omitempty"`
}

// 硬件信息
type Hardware struct {
    // 宿主机内存使用情况
    MemInfo  *utils.MemInfo  `json:"mem_stats,omitempty"`
    // 宿主机CPU使用情况
    CPUInfo  *utils.CPUInfo  `json:"cpu_stats,omitempty"`
    // 宿主机磁盘使用情况
    DiskInfo *utils.DiskInfo `json:"disk_stats,omitempty"`
    // 宿主机GPU信息和使用情况
    GPUInfo  []utils.GPUInfo `json:"gpu_stats,omitempty"`
}
```

#### /system/update

该接口用于更新系统中的应用，我们称之为应用OTA，后续还会实现主程序OTA（即OpenEdge主程序的自升级）。应用OTA会先停止所有老服务再启动所有新服务，所以有停服时间，后续会继续优化避免重启未更新的服务。一次应用OTA的过程如下：

![update](../../images/overview/design/openedge_update.png)

_**注意**：目前应用OTA采用全量更新的方式，即先停止所有老服务再启动所有新服务，因此服务会中断。_

#### /services/{}/instances/{}/start|stop

该接口用于动态启停某个服务的实例，需要指定服务名和实例名，如果重复启动同一个服务的相同名称的实例，会先把之前启动的实例停止，然后启动新的实例。

该接口可带上动态的配置用于覆盖存储卷中的静态配置，覆盖逻辑采用环境变量的方式，实例启动时可以加载环境变量的配置来覆盖存储卷中的配置，来避免资源冲突。比如进程模式下，函数计算的管理服务启动函数运行时服务时会预先分配空闲的端口，来使函数运行时实例监听于不同的端口。

### 环境变量

OpenEdge目前会给服务实例设置如下几个系统环境变量：

- OPENEDGE_HOST_OS：OpenEdge所在设备（宿主机）的系统类型
- OPENEDGE_MASTER_API：OpenEdge主程序的API Server地址
- OPENEDGE_RUNNING_MODE：OpenEdge主程序采用的服务运行模式
- OPENEDGE_SERVICE_NAME：被启动服务的名称
- OPENEDGE_SERVICE_TOKEN：动态分配的Token

官方提供的函数计算管理服务就是通过读取`OPENEDGE_MASTER_API`来连接OpenEdge主程序的，比如Linux系统下`OPENEDGE_MASTER_API`默认是 `unix:///var/run/openedge.sock`；其他系统的容器模式下`OPENEDGE_MASTER_API`默认是`tcp://host.docker.internal:50050`；其他系统的进程模式下`OPENEDGE_MASTER_API`默认是`tcp://127.0.0.1:50050`。

_**注意**：应用中配置的环境变量如果和上述系统环境变量相同会被覆盖。_

## 官方模块

目前官方提供了若干模块，用于满足部分常见的应用场景，当然开发者也可以开发自己的模块。

### openedge-agent

openedge-agent又称云代理模块，负责和BIE云端管理套件通讯，拥有MQTT和HTTPS通道，MQTT强制SSL/TLS证书双向认证，HTTPS强制SSL/TLS证书单向认证。开发者可以参考该模块实现自己的Agent模块来对接自己的云平台。

云代理目前就做两件事：

1. 启动后定时向主程序获取状态信息并上报给云端
2. 监听云端下发的事件，触发响应操作，目前只处理应用OTA事件

云代理接收到BIE云端管理套件的应用OTA指令后，会先下载所有配置中使用的存储卷数据包并解压到指定位置，如果存储卷数据包已经存在并且MD5相同则不会重复下载。所有存储卷都准备好之后，云代理模块会调用主程序的/update/system接口触发主程序更新系统。

_**提示**：如果设备无法连接外网或者需要脱离云端管理，可以从应用配置中移除Agent模块，离线运行。_

### openedge-hub

openedge-hub简称Hub是一个单机版的消息订阅和发布中心，采用MQTT3.1.1协议，可在低带宽、不可靠网络中提供可靠的消息传输服务。其作为OpenEdge系统的消息中间件，为所有服务提供消息驱动的互联能力。

目前支持4种接入方式：TCP、SSL（TCP + SSL）、WS(Websocket)及 WSS（Websocket + SSL），MQTT 协议支持度如下：

- 支持 Connect、Disconnect、Subscribe、Publish、Unsubscribe、Ping 等功能
- 支持 QoS 等级 0 和 1 的消息发布和订阅
- 支持 Retain、Will message、Clean Session
- 支持订阅含有 `+`、`#` 等通配符的主题
- 支持符合约定的 ClientID 和 Payload 的校验
- 暂时 **不支持** 发布和订阅以 `$` 为前缀的主题
- 暂时 **不支持** Client 的 Keep Alive 特性以及 QoS 等级 2 的发布和订阅

_**注意**：_

- 发布和订阅主题中含有的分隔符 `/` 最多不超过 8 个，主题名称长度最大不超过 255 个字符
- 消息报文默认最大长度位 `32k`，可支持的最大长度为 268,435,455(Byte)，约 256 MB，可通过 `message` 配置项进行修改
- ClientID 支持大小写字母、数字、下划线、连字符（减号）和空字符(如果CleanSession为false不允许为空), 最大长度不超过 128 个字符
- 消息的 QoS 只能降不能升，比如原消息的 QoS 为 0 时，即使订阅 QoS 为 1，消息仍然以 QoS 为 0 的等级发送
- 如果使用证书双向认证，Client必须在连接时发送非空的username和空的password，username会用于主题鉴权。如果password不为空，则还会进一步检查密码是否正确

Hub支持简单的主题路由，比如订阅主题为 `t` 的消息并以新主题 `t/topic` 发布回 Hub内部的broker。

如果该模块无法满足您的要求，您也可以使用第三方的MQTT Broker/Server来替换。

### openedge-function-manager

openedge-function-manager又称函数管理模块，提供基于MQTT消息机制，弹性、高可用、扩展性好、响应快的的计算能力，并且兼容[百度云-函数计算 CFC](https://cloud.baidu.com/product/cfc.html)。需要注意的是函数计算不保证消息顺序，除非只启动一个函数实例。

函数管理模块负责管理所有函数实例和消息路由规则，支持自动扩容和缩容。结构图如下：

![函数计算服务](../../images/overview/design/openedge_function.png)

如果函数执行错误，函数计算会返回如下格式的消息，供后续处理。其中packet是函数输入的消息（被处理的消息），不是函数返回的消息。示例如下：

```python
{
    "errorMessage": "rpc error: code = Unknown desc = Exception calling application",
    "errorType": "*errors.Err",
    "functionMessage": {
        "ID": 0,
        "QOS": 0,
        "Topic": "t",
        "Payload": "eyJpZCI6MSwiZGV2aWNlIjoiMTExIn0=",
        "FunctionName": "sayhi",
        "FunctionInvokeID": "50f8f102-2b8c-4904-86df-0728811a5a4b"
    }
}
```

### openedge-function-python27

openedge-function-python27提供Python函数与[百度云-函数计算 CFC](https://cloud.baidu.com/product/cfc.html)类似，用户通过编写的自己的函数来处理消息，可进行消息的过滤、转换和转发等，使用非常灵活。该模块可作为GRPC服务单独启动，也可以为函数管理模块提供函数运行实例。

Python函数的输入输出可以是JSON格式也可以是二进制形式。消息Payload在作为参数传给函数前会尝试一次JSON解码（`json.loads(payload)`），如果成功则传入字典（dict）类型，失败则传入原二进制数据。

Python函数支持读取环境变量，比如 os.environ['PATH']。

Python函数支持读取上下文，比如 context['functionName']。

Python函数实现举例：

```python
#!/usr/bin/env python
#-*- coding:utf-8 -*-
"""
module to say hi
"""

def handler(event, context):
    """
    function handler
    """
    event['functionName'] = context['functionName']
    event['functionInvokeID'] = context['functionInvokeID']
    event['messageQOS'] = context['messageQOS']
    event['messageTopic'] = context['messageTopic']
    event['sayhi'] = '你好，世界！'
    return event
```

_**提示**：Native进程模式下，若要运行本代码库example中提供的sayhi.py，需要自行安装python2.7，且需要基于python2.7安装protobuf3、grpcio (采用 pip 安装即可，`pip install grpcio protobuf`)。_

### openedge-remote-mqtt

openedge-remote-mqtt又称远程MQTT通讯模块，可桥接两个MQTT Server进行消息同步。目前支持配置多路消息转发，可配置多个Remote和Hub同时进行消息同步，结构图如下：

![远程MQTT通讯举例](../../images/overview/design/openedge_remote_mqtt.png)

如上图示，这里OpenEdge的本地Hub与远程云端Hub平台之间通过OpenEdge远程MQTT通讯模块实现消息的转发、同步，进一步地，通过在两端接入MQTT Client即可实现**端云协同式**的消息转发与传递。
