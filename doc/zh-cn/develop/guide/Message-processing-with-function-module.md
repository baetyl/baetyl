# 测试前准备

**声明**：本文测试所用设备系统为MacOS，模拟MQTT client行为的客户端为[MQTTBOX](http://workswithweb.com/html/mqttbox/downloads.html)。

与基于本地Hub模块实现设备间消息转发不同的是，本文在Hub模块的基础上，引入Function函数计算模块及具体执行所需的Python2.7 runtime模块，将接收到的消息交给Python2.7 runtime来处理（Python2.7 runtime会调用具体的函数脚本来执行具体计算、分析、处理等），然后将处理结果以主题方式反馈给Hub模块，最终订阅该主题的MQTT client（要求MQTT client事先订阅该主题）将会收到该处理结果。

本地Hub模块的配置项信息不再赘述，详情查看[基于Hub模块实现设备间消息转发](./Message-transfer-among-devices-with-hub-module.md)，这里主要介绍新引入的Function函数计算模块相关配置（Python函数的运行环境构建参考[OpenEdge Function模块设计](../../about/design/OpenEdge-function-module-design.md)），具体如下：

```yaml
name: [必须]模块名
hub:
  clientid: mqtt client连接hub的client id，如果为空则随机生成，且clean session强制变成true
  address: [必须]mqtt client连接hub的地址，docker容器模式下地址为hub模块名，native进程模式下为127.0.0.1
  username: 如果采用账号密码，必须填mqtt client连接hub的用户名
  password: 如果采用账号密码，必须填mqtt client连接hub的密码
  ca: 如果采用证书双向认证，必须填mqtt client连接hub的CA证书所在路径
  key: 如果采用证书双向认证，必须填mqtt client连接hub的客户端私钥所在路径
  cert: 如果采用证书双向认证，必须填mqtt client连接hub的客户端公钥所在路径
  timeout: 默认值：30s，mqtt client连接hub的超时时间
  interval: 默认值：1m，mqtt client连接hub的重连最大间隔时间，从500微秒翻倍增加到最大值。
  keepalive: 默认值：30s，mqtt client连接hub的保持连接时间
  cleansession: 默认值：false，mqtt client连接hub的clean session
  buffersize: 默认值：10，mqtt client发送消息给hub的内存队列大小，异常退出会导致消息丢失，恢复后QoS为1的消息依赖hub重发
rules: 路由规则配置项
  - id: [必须]路由规则ID
    subscribe:
      topic: [必须]向hub订阅的消息主题
      qos: 默认值：0，向hub订阅的消息QoS
    compute:
      function: [必须]处理消息的函数名
    publish:
      topic: [必须]函数处理输出结果消息发布到hub的主题
      qos: 默认值：0，函数处理输出结果消息发布到hub的QoS
functions:
  - name: [必须]函数名
    runtime: 配置函数依赖的runtime模块名称，python为'python2.7'
    entry: 同模块的entry，运行函数实例的runtime模块的镜像或可执行程序
    handler: [必须]函数处理函数。python为函数包和处理函数名，比如：'sayhi.handler'
    codedir: 如果是python，必须填python代码所在路径
    env: 环境变量配置项，例如：
      USER_ID: acuiot
    instance: 函数实例配置项
      min: 默认值：0，最小值：0，最大值：100，最小函数实例数
      max: 默认值：1，最小值：1，最大值：100，最大函数实例数
      timeout: 默认值：5m， 函数实例调用超时时间
      message:
        length:
          max: 默认值：4m， 函数实例允许接收和发送的最大消息长度
      cpu:
        cpus: 函数实例模块可用的CPU比例，例如：1.5，表示可以用1.5个CPU内核
        setcpus: 函数实例模块可用的CPU内核，例如：0-2，表示可以使用第0到2个CPU内核；0，表示可以使用第0个CPU内核；1，表示可以使用第1个CPU内核
      memory:
        limit: 函数实例模块可用的内存，例如：500m，表示可以用500兆内存
        swap: 函数实例模块可用的交换空间，例如：1g，表示可以用1G内存
      pids:
        limit: 函数实例模块可创建的进程数
```

_**提示**：凡是在rules消息路由配置项中出现、用到的函数，必须在functions配置项中进行函数执行具体配置，否则将不予启动。_

本文将以TCP连接方式为例，展示本地Function模块的消息处理、计算功能。

# 操作流程

- **Step1**：依据使用需求编写配置文件信息，然后以Docker容器模式启动OpenEdge可执行程序；
- **Step2**：通过MQTTBOX以TCP方式与OpenEdge Hub[建立连接](./Device-connect-with-OpenEdge-base-on-hub-module.md)；
    - 若成功与OpenEdge Hub模块建立连接，则依据配置的主题权限信息向有权限的主题发布消息，同时向拥有订阅权限的主题订阅消息，并观察OpenEdge日志信息；
      - 若OpenEdge日志显示已经启动Python runtime模块，则表明发布的消息受到了预期的函数处理；
      - 若OpenEdge日志显示未成功启动Python runtime模块，则重复上述操作，直至看到OpenEdge主程序成功启动了Python runtime模块。
    - 若与OpenEdge Hub建立连接失败，则重复**Step2**操作，直至MQTTBOX与OpenEdge Hub成功建立连接为止。
- **Step3**：通过MQTTBOX查看对应主题消息的收发状态。

![基于Function模块实现设备消息处理流程](../../images/develop/guide/process/openedge-python-flow.png)

# 消息路由测试

本文测试使用的本地Hub及Function模块的相关配置信息如下：

```yaml
# 本地Hub模块配置：
name: openedge_hub
mark: modu-nje2uoa9s
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

# 本地Function模块配置：
name: openedge_function
mark: modu-e1iluuach
hub:
  address: tcp://openedge_hub:1883
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
      topic: t/py
      qos: 1
functions:
  - name: 'sayhi'
    runtime: 'python2.7'
    handler: 'sayhi.handler'
    codedir: 'app/func-nyeosbbch'
    entry: "hub.baidubce.com/openedge-sandbox/openedge_function_runtime_python2.7:0.3.6"
    env:
      USER_ID: acuiot
    instance:
      min: 1
      max: 10
      timeout: 30s
```

如上配置，假若MQTTBOX基于上述配置信息已与本地Hub模块建立连接，向主题“t”发送的消息将会交给“sayhi”函数处理，然后将处理结果以主题“t/py”发布回Hub模块，这时订阅主题“t/py”的MQTT client将会接收到这条处理后的消息。

## OpenEdge 启动

如**Step1**所述，以Docker容器模式启动OpenEdge，通过观察OpenEdge启动日志可以发现Hub、Function模块均已被成功加载，具体如下图示。

![OpenEdge加载、启动日志](../../images/develop/guide/process/openedge-function-start.png)

同样，我们也可以通过执行命令`docker ps`查看系统当前正在运行的docker容器列表，具体如下图示。

![通过`docker ps`命令查看系统当前运行docker容器列表](../../images/develop/guide/process/openedge-docker-ps-after.png)

经过对比，不难发现，本次OpenEdge启动时已经成功加载了Hub、Function两个容器模块。

## MQTTBOX 建立连接

本次测试中，我们采用TCP连接方式对MQTTBOX进行连接信息配置，然后点击“Add subscriber”按钮订阅主题“t/py”，该主题用于接收经python函数“sayhi”处理之后的结果数据，具体如下图示。

![MQTTBOX连接配置](../../images/develop/guide/process/mqttbox-tcp-process-config.png)

上图显示，MQTTBOX已经成功订阅了主题“t/py”。

## 消息路由验证

根据上文所述，这里我们利用python函数“sayhi”对主题“t”的消息进行处理，并将结果反馈给主题“t/py”。那么，首先，需要获悉的就是处理函数“sayhi.py”的具体信息，具体如下示：

```python
#!/usr/bin/env python
#-*- coding:utf-8 -*-
"""
module to say hi
"""

import os
import time
import threading


def handler(event, context):
    """
    function handler
    """

    event['USER_ID'] = os.environ['USER_ID']
    event['functionName'] = context['functionName']
    event['functionInvokeID'] = context['functionInvokeID']
    event['invokeid'] = context['invokeid']
    event['messageQOS'] = context['messageQOS']
    event['messageTopic'] = context['messageTopic']
    event['py'] = '你好，世界！'

    return event
```

可以发现，在接收到某Json格式的消息后，函数“sayhi.py”会对其进行一系列处理，然后将处理结果返回。返回的结果中包括：环境变量“USER_ID”、函数名称“functionName”、函数调用ID“functionInvokeID”、输入消息主题“messageTopic”、输入消息消息QoS“messageQOS”等字段。

这里，我们通过MQTTBOX将消息“{"id":10}”发布给主题“t”，然后观察主题“t/py”的接收消息情况，具体如下图示。

![MQTTBOX成功接收到经python函数处理之后的消息](../../images/develop/guide/process/mqttbox-tcp-process-success.png)，且结果与上面的分析结果吻合。由此，我们完成了消息路由的处理测试。

此外，我们这时可以观察OpenEdge的日志及再次执行命令`docker ps`查看系统当前正在运行的容器列表，其结果如下图示。

![运用python函数处理消息时OpenEdge日志](../../images/develop/guide/process/openedge-python-start.png)

![通过docker ps命令查看系统当前正在运行的容器列表](../../images/develop/guide/process/openedge-docker-ps-python-start.png)

从上述两张图片中可以看出，除了OpenEdge启动时已加载的Hub、Function模块容器，在利用python函数“sayhi.py”对主题“t”消息进行处理时，系统还启动、并运行了Python Runtime模块，其主要用于对消息作运行时处理（各类模块加载、启动细节可参见[OpenEdge主程序模块设计](../../about/design/OpenEdge-master-module-design.md)）。