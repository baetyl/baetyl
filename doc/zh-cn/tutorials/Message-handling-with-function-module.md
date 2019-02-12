# 利用本地函数计算模块进行消息处理

**声明**：

> + 本文测试所用设备系统为 Darwin
> + 模拟 MQTT client 行为的客户端为 [MQTTBOX](../Resources-download.md#下载MQTTBOX客户端)
> + 本文所用镜像为依赖 OpenEdge 源码自行编译所得，具体请查看[如何从源码构建镜像](../setup/Build-OpenEdge-from-Source.md)

与基于本地 Hub 模块实现设备间消息转发不同的是，本文主要介绍利用本地函数计算模块进行消息处理。其中，本地 Hub 模块用于建立 OpenEdge 与 MQTT 客户端之间的连接，Python 运行时模块用于处理 MQTT 消息，而本地函数计算模块则通过 MQTT 消息上下文衔接本地 Hub 模块与 Python 运行时模块。

本文将以 TCP 连接方式为例，展示本地函数计算模块的消息处理、计算功能。

## 操作流程

- Step 1：以 Docker 容器模式启动 OpenEdge 可执行程序；
- Step 2：通过 MQTTBOX 以 TCP 方式与 OpenEdge Hub 模块[建立连接](./Device-connect-to-OpenEdge-with-hub-module.md)；
    - 若成功与 OpenEdge Hub 模块建立连接，则依据配置的主题权限信息向有权限的主题发布消息，同时向拥有订阅权限的主题订阅消息，并观察 OpenEdge 日志信息；
      - 若 OpenEdge 日志显示已经启动 Python 运行时模块，则表明发布的消息受到了预期的函数处理；
      - 若 OpenEdge 日志显示未成功启动 Python 运行时模块，则重复上述操作，直至看到 OpenEdge 主程序成功启动了 Python 运行时模块。
    - 若与 OpenEdge Hub 建立连接失败，则重复 `Step 2` 操作，直至 MQTTBOX 与 OpenEdge Hub 模块成功建立连接为止。
- Step 3：通过 MQTTBOX 查看对应主题消息的收发状态。

![基于本地函数计算模块实现设备消息处理流程](../../images/tutorials/process/openedge-python-flow.png)

## 消息处理测试

本文测试使用的本地 Hub 及函数计算模块的相关配置信息如下：

```yaml
name: localhub
listen:
  - tcp://:1883
principals:
  - username: 'test'
    password: 'be178c0543eb17f5f3043021c9e5fcf30285e557a4fc309cce97ff9ca6182912'
    permissions:
      - action: 'pub'
        permit: ['#']
      - action: 'sub'
        permit: ['#']

# 本地函数计算模块配置：
name: localfunc
hub:
  address: tcp:/hub:1883
  username: test
  password: hahaha
rules:
  - id: rule-e1iluuac1
    subscribe:
      topic: t
      qos: 1
    compute:
      function: sayhi
    publish:
      topic: t/hi
      qos: 1
functions:
  - id: func-nyeosbbch
    name: 'sayhi'
    runtime: 'python27'
    handler: 'sayhi.handler'
    codedir: 'var/db/openedge/module/func-nyeosbbch'
    entry: "openedge-function-runtime-python27:build"
    env:
      USER_ID: acuiot
    instance:
      min: 0
      max: 10
      timeout: 1m
```

如上配置，假若 MQTTBOX 基于上述配置信息已与本地 Hub 模块建立连接，向主题 `t` 发送的消息将会交给 `sayhi` 函数处理，然后将处理结果以主题 `t/hi` 发布回 Hub 模块，这时订阅主题 `t/hi` 的 MQTT client 将会接收到这条处理后的消息。

_**提示**：凡是在 `rules` 消息路由配置项中出现、用到的函数，必须在 `functions` 配置项中进行函数执行具体配置，否则将不予启动。_

### OpenEdge 启动

如 `Step 1` 所述，以 Docker 容器模式启动 OpenEdge，通过观察 OpenEdge 启动日志可以发现本地 Hub 模块和函数计算模块均已被成功加载，具体如下图示。

![OpenEdge 加载、启动日志](../../images/tutorials/process/openedge-function-start.png)

同样，我们也可以通过执行命令`docker ps`查看系统当前正在运行的docker容器列表，具体如下图示。

![通过 `docker ps` 命令查看系统当前运行 Docker 容器列表](../../images/tutorials/process/openedge-docker-ps-after.png)

经过对比，不难发现，本次 OpenEdge 启动时已经成功加载了本地 Hub 模块和函数计算模块两个容器模块。

### MQTTBOX 建立连接

本次测试中，我们采用 TCP 连接方式对 MQTTBOX 进行连接信息配置，然后点击 `Add subscriber` 按钮订阅主题 `t/hi` ，该主题用于接收经 Python 函数 `sayhi` 处理之后的结果数据，具体如下图示。

![MQTTBOX 连接配置](../../images/tutorials/process/mqttbox-tcp-process-config.png)

上图显示，MQTTBOX 已经成功订阅了主题 `t/hi` 。

### 消息处理验证

根据上文所述，这里我们利用 Python 函数 `sayhi` 对主题 `t` 的消息进行处理，并将结果反馈给主题 `t/hi` 。那么，首先，需要获悉的就是处理函数 `sayhi` 的具体信息，具体如下示：

```python
#!/usr/bin/env python
#-*- coding:utf-8 -*-
"""
module to say hi
"""

import os


def handler(event, context):
    """
    function handler
    """
    if 'USER_ID' in os.environ:
      event['USER_ID'] = os.environ['USER_ID']

    if 'functionName' in context:
      event['functionName'] = context['functionName']

    if 'functionInvokeID' in context:
      event['functionInvokeID'] = context['functionInvokeID']

    if 'invokeid' in context:
      event['invokeid'] = context['invokeid']

    if 'messageQOS' in context:
      event['messageQOS'] = context['messageQOS']

    if 'messageTopic' in context:
      event['messageTopic'] = context['messageTopic']

    event['py'] = '你好，世界！'

    return event
```

可以发现，在接收到某字典类格式的消息后，函数 `sayhi` 会对其进行一系列处理，然后将处理结果返回。返回的结果中包括：环境变量 `USER_ID`、函数名称 `functionName`、函数调用 ID `functionInvokeID`、输入消息主题 `messageTopic`、输入消息消息 QoS `messageQOS` 等字段。

这里，我们通过 MQTTBOX 将消息 `{"id":10}` 发布给主题 `t` ，然后观察主题 `t/hi` 的接收消息情况，具体如下图示。

![MQTTBOX 成功接收到经 Python 函数处理之后的消息](../../images/tutorials/process/mqttbox-tcp-process-success.png)，且结果与上面的分析结果吻合。由此，我们完成了消息路由的处理测试。

此外，我们这时可以观察 OpenEdge 的日志及再次执行命令 `docker ps` 查看系统当前正在运行的容器列表，其结果如下图示。

![运用 Python 函数处理消息时 OpenEdge 日志](../../images/tutorials/process/openedge-python-start.png)

![通过 `docker ps` 命令查看系统当前正在运行的容器列表](../../images/tutorials/process/openedge-docker-ps-python-start.png)

从上述两张图片中可以看出，除了 OpenEdge 启动时已加载的本地 Hub 模块和函数计算模块容器，在利用 Python 函数 `sayhi` 对主题 `t` 消息进行处理时，系统还启动、并运行了 Python 运行时模块，其主要用于对消息作运行时处理（各类模块加载、启动细节可参见 [OpenEdge 设计](../overview/OpenEdge-design.md)）。
