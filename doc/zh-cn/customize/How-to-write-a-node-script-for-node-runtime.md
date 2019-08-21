# 如何针对 Node 运行时编写 js 脚本

**声明**：

- 本文测试所用设备系统为 Ubuntu18.04
- node 版本为 8.5
- 模拟 MQTT client 行为的客户端为 [MQTTBOX](../Resources-download.md#下载-MQTTBOX-客户端)
- 本文中基于 Hub 模块创建的服务名称为 `localhub` 服务。并且针对本文的测试案例中，对应的 `localhub` 服务、函数计算服务以及其他服务的配置统一如下：

```yaml
# 本地 Hub 配置
# 配置文件位置: var/db/openedge/localhub-conf/service.yml
listen:
  - tcp://0.0.0.0:1883
principals:
  - username: 'test'
    password: 'hahaha'
    permissions:
      - action: 'pub'
        permit: ['#']
      - action: 'sub'
        permit: ['#']

# 本地 openedge-function-manager 配置
# 配置文件位置: var/db/openedge/function-manager-conf/service.yml
hub:
  address: tcp://localhub:1883
  username: test
  password: hahaha
rules:
  - clientid: localfunc-1
    subscribe:
      topic: node
    function:
      name: sayhi
    publish:
      topic: t/hi
functions:
  - name: sayhi
    service: function-sayhi
    instance:
      min: 0
      max: 10
      idletime: 1m

# node function 配置
# 配置文件位置: var/db/openedge/function-sayjs-conf/service.yml
functions:
  - name: 'sayhi'
    handler: 'index.handler'
    codedir: 'var/db/openedge/function-sayhi'

# application.yml配置
# 配置文件位置: var/db/openedge/application.yml
version: v0
services:
  - name: localhub
    image: openedge-hub
    replica: 1
    ports:
      - 1883:1883
    mounts:
      - name: localhub-conf
        path: etc/openedge
        readonly: true
      - name: localhub-data
        path: var/db/openedge/data
      - name: localhub-log
        path: var/log/openedge
  - name: function-manager
    image: openedge-function-manager
    replica: 1
    mounts:
      - name: function-manager-conf
        path: etc/openedge
        readonly: true
      - name: function-manager-log
        path: var/log/openedge
  - name: function-sayhi
    image: openedge-function-node85
    replica: 0
    mounts:
      - name: function-sayjs-conf
        path: etc/openedge
        readonly: true
      - name: function-sayjs-code
        path: var/db/openedge/function-sayhi
        readonly: true
volumes:
  # hub
  - name: localhub-conf
    path: var/db/openedge/localhub-conf
  - name: localhub-data
    path: var/db/openedge/localhub-data
  - name: localhub-log
    path: var/db/openedge/localhub-log
  # function manager
  - name: function-manager-conf
    path: var/db/openedge/function-manager-conf
  - name: function-manager-log
    path: var/db/openedge/function-manager-log
  # function node runtime sayhi
  - name: function-sayjs-conf
    path: var/db/openedge/function-sayjs-conf
  - name: function-sayjs-code
    path: var/db/openedge/function-sayjs-code
```

OpenEdge 官方提供了 Node 运行时，可以加载用户所编写的 js 脚本。下文将针对 js 脚本的名称，执行函数名称，输入，输出参数等内容分别进行说明。

## 函数名约定

js 脚本的名称可以参照 js 的通用命名规范，OpenEdge 并未对此做特别限制。如果要应用某 js 脚本对某条 MQTT 消息做处理，则相应的函数运行时服务的配置如下：

```yaml
functions:
  - name: 'sayhi'
    handler: 'index.handler'
    codedir: 'var/db/openedge/function-sayhi'
```

这里，我们关注 `handler` 这一属性，其中 `index` 代表脚本名称，后面的 `handler` 代表该文件中被调用的入口函数。

```
function-sayjs-code/
└── index.js
```

更多函数运行时服务配置请查看 [函数运行时服务配置释义](../tutorials/Config-interpretation.md)。

## 参数约定

```javascript
exports.handler = (event, context, callback) => {
    callback(null, event);
};
```

OpenEdge 官方提供的 Node 运行时支持 2 个参数: event 和 context，下面将分别介绍其用法。

- **event**：根据 MQTT 报文中的 Payload 传入不同参数
    - 若原始 Payload 为一个 Json 数据，则传入经过 json.loads(Payload) 处理后的数据;
    - 若原始 Payload 为字节流、字符串(非 Json)，则传入原 Payload 数据。
- **context**：MQTT 消息上下文
    - context.messageQOS // MQTT QoS
    - context.messageTopic // MQTT Topic
    - context.functionName // MQTT functionName
    - context.functionInvokeID //MQTT function invokeID
    - context.invokeid // 同上，用于兼容 [CFC](https://cloud.baidu.com/product/cfc.html)

_**提示**：在云端 CFC 测试时，请注意不要直接使用 OpenEdge 定义的上下文信息。推荐做法是先判断字段是否在 context 中存在，如果存在再读取。_

## Hello World!

下面我们实现一个简单的 js 脚本，目标是为每一条流经需要用该 js 脚本进行处理的 MQTT 消息附加一条 `hello world` 信息。对于字典类消息，将其直接返回即可，对于非字典类消息，则将之转换为字符串后返回。

```javascript
#!/usr/bin/env node

exports.handler = (event, context, callback) => {
  result = {};
  
  if (Buffer.isBuffer(event)) {
      const message = event.toString();
      result["msg"] = message;
      result["type"] = 'non-dict';
  }else {
      result["msg"] = event;
      result["type"] = 'dict';
  }

  result["say"] = 'hello world';
  callback(null, result);
};
```

+ **发送字典类数据**:

![发送字典类数据](../../images/customize/write-node-script-dict.png)

+ **发送非字典类数据**:

![发送非字典类数据](../../images/customize/write-node-script-none-dict.png)

如上，对于一些常规的需求，我们通过系统 Node 环境的标准库就可以完成。但是，对于一些较为复杂的需求，往往需要引入第三方库来完成。如何解决这个问题？我们将在 [如何针对 Node 运行时引入第三方包](./How-to-import-third-party-libraries-for-node-runtime.md) 小节详述。
