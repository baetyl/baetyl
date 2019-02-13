# 如何针对 Python 运行时编写 Python 脚本

**声明**：

> + 本文测试所用设备系统为 Darwin
> + 模拟 MQTT client 行为的客户端为 [MQTTBOX](../Resources-download.md#下载MQTTBOX客户端)
> + 本文所提到的测试案例中，对应本地 Hub 模块和函数计算模块的配置统一配置如下

```yaml
# 本地 Hub 模块配置：
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
  address: tcp://localhub:1883
  username: test
  password: hahaha
rules:
  - id: rule-e1iluuac1
    subscribe:
      topic: py
      qos: 0
    compute:
      function: get
    publish:
      topic: py/hi
      qos: 0
functions:
  - id: func-nyeosbbch
    name: 'get'
    runtime: 'python27'
    handler: 'get.handler'
    codedir: 'var/db/openedge/module/func-nyeosbbch'
    entry: "hub.baidubce.com/openedge/openedge-function-runtime-python27:0.1.1"
```

OpenEdge 官方提供了 Python 运行时，可以加载用户所编写的 Python 脚本。下文将针对 Python 脚本的名称，执行函数名称，输入，输出参数等内容分别进行说明。

## 函数名约定

Python 脚本的名称可以参照 Python 的通用命名规范，OpenEdge 并未对此做特别限制。如果要应用某 Python 脚本对某条 MQTT 消息做处理，则相应的函数计算模块的配置如下：

```yaml
functions:
  - id: func-nyeosbbch
    name: 'sayhi'
    runtime: 'python27'
    handler: 'sayhi.handler'
    codedir: 'var/db/openedge/module/func-nyeosbbch'
    entry: "hub.baidubce.com/openedge/openedge-function-runtime-python27:0.1.1"
```

这里，我们关注 `handler` 这一属性，其中 `sayhi` 代表脚本名称，后面的 `handler` 代表该文件中被调用的入口函数。

```
func-nyeosbbch
    sayhi.py 
```

更多函数模块配置请查看[函数计算模块配置释义](../tutorials/Config-interpretation.md)。

## 参数约定

```python
def handler(event, context):
    # do something
    return event
```

OpenEdge 官方提供的 Python 运行时支持2个参数: event 和 context，下面将分别介绍其用法。

+ **event**：根据 MQTT 报文中的 Payload 传入不同参数
    + 若原始 Payload 为一个 Json 数据，则传入经过 json.loads(Payload) 处理后的数据;
    + 若原始 Payload 为字节流、字符串(非 Json)，则传入原 Payload 数据。
+ **context**：MQTT 消息上下文
    + context.messageQOS // MQTT QoS
    + context.messageTopic // MQTT Topic
    + context.functionName // MQTT functionName
    + context.functionInvokeID //MQTT function invokeID
    + context.invokeid // 同上，用于兼容 [CFC](https://cloud.baidu.com/product/cfc.html)

_**提示**：在云端 CFC 测试时，请注意不要直接使用 OpenEdge 定义的上下文信息。推荐做法是先判断字段是否在 context 中存在，如果存在再读取。_

## Hello World!

下面我们实现一个简单的 Python 函数，目标是为每一条流经需要用该 Python 脚本进行处理的 MQTT 消息附加一条 `hello world` 信息。对于字典类消息，将其直接返回即可，对于非字典类消息，则将之转换为字符串后返回。

```python
#!/usr/bin/env python
# -*- coding: utf-8 -*-

def handler(event, context):
    result = {}
    if isinstance(event, dict):
        result['msg'] = event
        result['type'] = 'dict'
        result['say'] = 'hello world'
    else:
        result['msg'] = event
        result['type'] = 'non-dict'
        result['say'] = 'hello world'

    return result
```

+ **发送字典类数据**：

![发送字典类数据](../../images/customize/write-python-script-dict.png)

+ **发送非字典类数据**：

![发送非字典类数据](../../images/customize/write-python-script-none-dict.png)

如上，对于一些常规的需求，我们通过系统 Python 环境的标准库就可以完成。但是，对于一些较为复杂的需求，往往需要引入第三方库来完成。如何解决这个问题？我们将在下文详述。

## 如何引用第三方包

通常情况下，系统自带的 Python 环境很可能不会满足我们的需要，实际使用往往需要引入第三方库，下面给出一个示例。

假定我们想要对一个网站进行爬虫，获取相应的信息。这里，我们可以引入第三方库 [requests](https://pypi.org/project/requests)。如何引入，具体如下所示：

> + 步骤1: `pip download requests` // 下载 requests 及其依赖（idna、urllib3、chardet、certifi）
> + 步骤2: `cp requests-package /directory/to/Python/script` // 将下载的 requests 及其依赖的源码包拷贝到该 Python 脚本目录
> + 步骤3: `touch __init__.py` // 使执行脚本所在目录成为一个 package
> + 步骤4: `import requests` // 引入第三方库 requests，然后编写具体执行脚本
> + 步骤5: `python your_script.py` // 执行脚本

如上述操作正常，则形成的脚本目录结构如下图所示。

![Python 第三方库脚本目录](../../images/customize/python-third-lib-dir.png)

下面，我们编写脚本 `get.py` 来获取 [https://openedge.tech](https://openedge.tech) 的 headers 信息，假定触发条件为 Python 运行时接收到来自 Hub 的 `A` 指令，具体如下：

```python
#!/usr/bin/env python
# -*- coding: utf-8 -*-

import requests

def handler(event, context):
    """
    data: {"action": "A"}
    """
	if 'action' in event:
		if event['action'] == 'A':
			r = requests.get('https://openedge.tech')
			if str(r.status_code) == '200':
				event['info'] = r.headers
			else:
				event['info'] = 'exception found'
		else:
			event['info'] = 'action error'

	else:
		event['error'] = 'action not found'

	return event
```

如上，Hub 接收到发送到主题 `py` 的消息后，会调用 `get.py` 脚本执行具体处理逻辑，然后将执行结果以 MQTT 消息形式反馈给主题 `py/hi`。这里，我们通过 MQTTBOX 订阅主题 `py/hi`，并向主题 `py` 发送消息 `{"action": "A"}`，然后观察 MQTTBOX 订阅主题 `py/hi` 的消息收取情况，如正常，则可正常获取 [https://openedge.tech](https://openedge.tech) 的 headers 信息。

![获取OpenEdge官网headers信息](../../images/customize/write-python-script-third-lib.png)