# Baetyl 配置文件释义

- [说在前头](#说在前头)
- [主程序配置](#主程序配置)
- [应用配置](#应用配置)
- [baetyl-agent 配置](#baetyl-agent-配置)
- [baetyl-hub 配置](#baetyl-hub-配置)
- [baetyl-function-manager 配置](#baetyl-function-manager-配置)
- [baetyl-function-python27、baetyl-function-python36 配置](#baetyl-function-python27-配置-baetyl-function-python36-配置)
- [baetyl-remote-mqtt 配置](#baetyl-remote-mqtt-配置)

## 说在前头

支持的单位：

- 大小：b：字节（byte）；k：千字节（kilobyte）；m：兆字节（megabyte），g：吉字节（gigabyte）
- 时间：s：秒；m：分；h：小时

配置示例可参考本项目 `example` 目录中的各种例子。

## 主程序配置

主程序的配置和应用配置是分离的，默认配置文件是工作目录下的`etc/baetyl/baetyl.yml`，配置解读如下：

```yaml
mode: 默认值：docker，服务运行模式。docker：容器模式；native：进程模式
grace: 默认值：30s，服务优雅退出超时时间
server: 主程序API Server配置项
  address: 默认值可读取环境变量：OPENEDGE_MASTER_API，主程序 API Server 地址
  timeout: 默认值：30s，主程序 API Server 请求超时时间
logger: 日志配置项
  path: 默认为空，即不打印到文件；如果指定文件则输出到文件
  level: 默认值：info，日志等级，支持 debug、info、warn 和 error
  format: 默认值：text，日志打印格式，支持 text 和 json
  age:
    max: 默认值：15，日志文件保留的最大天数
  size:
    max: 默认值：50，日志文件大小限制，单位MB
  backup:
    max: 默认值：15，日志文件保留的最大数量
```

## 应用配置

应用配置的默认配置文件是工作目录下的 `var/db/baetyl/application.yml`，配置解读如下：

```yaml
version: 应用版本
services: 应用的服务列表
  - name: [必须]服务名称，在服务列表中必须唯一
    image: [必须]服务入口。Docker 容器模式下表示服务镜像；Native 进程模式下表示服务运行包所在位置
    replica: 默认为0，服务副本数，表示启动的服务实例数。通常服务只需启动一个。函数运行时服务一般设置为0，不由主程序启动，而是由函数管理服务来动态启动实例
    mounts: 存储卷映射列表
      - name: [必须]存储卷名称，对应存储卷列表中的一个
        path: [必须]存储卷映射到容器中的路径
        readonly: 默认值：false，存储卷是否只读
    ports: Docker 容器模式下暴露的端口，例如：
      - 0.0.0.0:1883:1883
      - 0.0.0.0:1884:1884/tcp
      - 8080:8080/tcp
      - 9884:8884
    devices: Docker 容器模式下的设备映射，例如：
      - /dev/video0
      - /dev/sda:/dev/xvdc:r
    args: 服务实例启动参数，例如：
      - '-c'
      - 'conf/conf.yml'
    env: 服务实例环境变量，例如：
      version: v1
    restart: 服务实例重启策略配置项
      retry:
        max: 默认为空，表示总是重试，服务重启最大次数
      policy: 默认值：always，重启策略。no：不重启；always：总是重启；on-failure：服务异常退出就重启
      backoff:
        min: 默认值：1s，重启最小间隔时间
        max: 默认值：5m，重启最大间隔时间
        factor: 默认值：2，重启间隔增大倍数
    resources: Docker 容器模式下的服务实例资源限制配置项
      cpu:
        cpus: 服务实例可用的 CPU 比例，例如：1.5，表示可以用1.5个 CPU 内核
        setcpus: 服务实例可用的 CPU 内核，例如：0-2，表示可以使用第0到2个 CPU 内核；0，表示可以使用第0个 CPU 内核；1，表示可以使用第1个 CPU 内核
      memory:
        limit: 服务实例可用的内存，例如：500m，表示可以用500兆内存
        swap: 服务实例可用的交换空间，例如：1g，表示可以用1G内存
      pids:
        limit: 服务实例可创建的进程数
volumes: 存储卷列表
  - name: [必须]存储卷名称，在存储卷列表中唯一
    path: [必须]存储卷在宿主机上的路径，相对于主程序的工作目录而言
```

## baetyl-agent 配置

```yaml
remote: Agent 模块对接 BIE 云端管理套件的配置项
  mqtt: BIE 云端 MQTT 通道配置
    clientid: [必须]连接云端 MQTT 通道的 Client ID，必须是云端核心设备的 ID
    address: [必须]云端 MQTT 通道的的地址，必须使用 SSL Endpoint
    username: [必须]云端 MQTT 通道连接的用户名，必须是云端核心设备的用户名
    ca: [必须]云端 MQTT 通道连接的CA证书路径
    key: [必须]云端 MQTT 通道连接的客户端私钥路径
    cert: [必须]云端 MQTT 通道连接的客户端公钥路径
    timeout: 默认值：30s，云端 MQTT 通道连接超时时间
    interval: 默认值：1m，云端 MQTT 通道重连的最大间隔时间，从500微秒翻倍增加到最大值
    keepalive: 默认值：1m，云端 MQTT 通道连接的保持时间
    cleansession: 默认值：false，云端 MQTT 通道连接的的是否保持 Session
    validatesubs: 默认值：false，云端 MQTT 通道连接是否检查订阅结果。 如果为 true 发现订阅失败报错退出
    buffersize: 默认值：10，发送消息内存队列大小，异常退出会导致消息丢失
  http: BIE 云端 HTTP 通道配置
    address: 会自动根据 MQTT 通道的地址推断云端 HTTPS 通道地址，无需配置
    timeout: 默认值：30s，云端HTTPS通道连接超时时间
  report: Agent 上报云端配置
    url: 上报的URL，无需配置
    topic: 上报主题模板，无需配置
    interval: 默认值：20s，上报间隔时间
  desire: Agent 接收云端下发配置
    topic: 下发主题模板，无需配置
```

## baetyl-hub 配置

```yaml
listen: [必须]监听地址，例如：
  - tcp://0.0.0.0:1883
  - ssl://0.0.0.0:1884
  - ws://:8080/mqtt
  - wss://:8884/mqtt
certificate: SSL/TLS 证书认证配置项，如果启用 ssl 或 wss 必须配置
  ca: Server 的 CA 证书路径
  key: Server 的服务端私钥路径
  cert: Server 的服务端公钥路径
principals: 接入权限配置项，如果不配置则 Client 无法接入，支持账号密码和证书认证
  - username: Client 接入 Hub 用户名
    password: Client 接入 Hub 密码
    permissions:
      - action: 操作权限。pub：发布权限；sub：订阅权限
        permit: 操作权限允许的主题列表，支持+和#匹配符
  - username: Client 接入 Hub 用户名，使用证书双向认证可不配置密码
    permissions:
      - action: 操作权限。pub：发布权限；sub：订阅权限
        permit: 操作权限允许的主题列表，支持+和#匹配符
subscriptions: 主题路由配置项
  - source:
      topic: 订阅的主题
      qos: 订阅的 QoS
    target:
      topic: 发布的主题
      qos: 发布的 QoS
message: 消息相关的配置项
  length:
    max: 默认值：32k；最大值：268,435,455字节(约256MB)，可允许传输的最大消息长度
  ingress: 消息接收配置项
    qos0:
      buffer:
        size: 默认值：10000，可缓存到内存中的 QoS 为0的消息数，增大缓存可提高消息接收的性能，若设备掉电，则会直接丢弃 QoS 为 0 的消息
    qos1:
      buffer:
        size:  默认值：100，等待持久化的 QoS 为 1 的消息缓存大小，增大缓存可提高消息接收的性能，但潜在的风险是 Hub 异常退出（比如设备掉电）会丢失缓存的消息，不回复确认（puback）。Hub 正常退出会等待缓存的消息处理完，不会丢失数据。
      batch:
        max:  默认值：50，批量写QoS为1的消息到数据库（持久化）的最大条数，消息持久化成功后会回复确认（ack）
      cleanup:
        retention:  默认值：48h，QoS 为 1 的消息保存在数据库中的时间，超过该时间的消息会在清理时物理删除
        interval:  默认值：1m，QoS 为 1 的消息清理时间间隔
  egress: 消息发送配置项
    qos0:
      buffer:
        size:  默认值：10000，内存缓存中的待发送的 QoS 为 0 的消息数，若设备掉电会直接丢弃消息；缓存满后，新推送的消息直接丢弃
    qos1:
      buffer:
        size:  默认值：100，QoS 为 1 的消息发送后，未确认（ack）的消息缓存大小，缓存满后，不再读取新消息，一直等待缓存中的消息被确认。QoS 为 1 的消息发送给客户端成功后等待客户端确认（puback），如果客户端在规定时间内没有回复确认，消息会一直重发，直到客户端回复确认或者 session 关闭
      batch:
        max:  默认值：50，批量从数据库读取消息的最大条数
      retry:
        interval:  默认值：20s，消息重发时间间隔
  offset: 消息序列号持久化相关配置
    buffer:
      size:  默认值：10000，被确认（ack）的消息的序列号的缓存队列大小。比如当前批量发送了QoS为1且序列号为1、2和3的三条消息给客户端，客户端确认了序列号1和3的消息，此时序列号1会入列并持久化，序列号3虽然已经确认，但是还是得等待序列号2被确认入列后才能入列。该设计可保证Hub异常重启后仍能从持久化的序列号恢复消息处理，保证消息不丢，但是会出现消息重发，也因此暂不支持QoS为2的消息
    batch:
      max:  默认值：100，批量写消息序列号到数据库的最大条数
logger: 日志配置项
  path: 默认为空，即不打印到文件；如果指定文件则输出到文件
  level: 默认值：info，日志等级，支持debug、info、warn 和 error
  format: 默认值：text，日志打印格式，支持 text 和 json
  age:
    max: 默认值：15，日志文件保留的最大天数
  size:
    max: 默认值：50，日志文件大小限制，单位MB
  backup:
    max: 默认值：15，日志文件保留的最大数量
status: Hub 状态配置项
  logging:
    enable: 默认值：false，是否打印Hub的状态信息
    interval: 默认值：1m，状态信息打印时间间隔
storage: 数据库存储配置项
  dir: 默认值：`var/db/baetyl/data`，数据库存储目录
shutdown: Hub 退出配置项
  timeout: 默认值：10m，Hub退出超时时间
```

## baetyl-function-manager 配置

```yaml
hub:
  clientid: Client 连接 Hub 的 Client ID。cleansession 为 false 则不允许为空
  address: [必须] Client 连接Hub的地址
  username: [必须] Client 连接Hub的用户名
  password: 如果采用账号密码，必须填 Client 连接Hub的密码，否者不用填写
  ca: 如果采用证书双向认证，必须填 Client 连接Hub的CA证书路径
  key: 如果采用证书双向认证，必须填 Client 连接Hub的客户端私钥路径
  cert: 如果采用证书双向认证，必须填 Client 连接Hub的客户端公钥路径
  timeout: 默认值：30s，Client 连接 Hub 的超时时间
  interval: 默认值：1m，Client 连接 Hub 的重连最大间隔时间，从500微秒翻倍增加到最大值
  keepalive: 默认值：1m，Client 连接Hub的保持连接时间
  cleansession: 默认值：false，Client 连接 Hub 的是否保持 Session
  validatesubs: 默认值：false，Client 是否检查 Hub 订阅结果，如果是发现订阅失败报错退出
  buffersize: 默认值：10，Client 发送消息给 Hub 的内存队列大小，异常退出会导致消息丢失，恢复后 QoS 为1的消息依赖 Hub 重发
rules: 路由规则配置项
  - clientid: Client 连接 Hub 的 Client ID
    subscribe:
      topic: [必须] Client 向 Hub 订阅的消息主题
      qos: 默认值：0，Client向 Hub 订阅的消息 QoS
    function:
      name: [必须]处理消息的函数名
    publish:
      topic: [必须]计算结果发布到 Hub 的主题
      qos: 默认值：0，计算结果发布 Hub 的 QoS
    retry:
      max: 默认值：3，最大重试次数
functions: 函数列表
  - name: [必须]函数名称，列表内唯一
    service: [必须]提供函数实例的服务名称
    instance: 实例配置项
      min: 默认值：0，最小值：0，最大值：100，最少实例数
      max: 默认值：1，最小值：1，最大值：100，最大实例数
      idletime: 默认值：10m，实例最大空闲时间
      evicttime: 默认值：1m，实例检查周期，如果发现实例空闲超过就销毁
    message:
      length:
        max: 默认值：4m， 函数实例允许接收和发送的最大消息长度
    backoff:
      max: 默认值：1m，Client 连接函数实例最大重连间隔
    timeout: 默认值：30s，Client 连接函数实例超时时间
```

## baetyl-function-python27、baetyl-function-python36 配置

```yaml
# 两个模块的配置方式相同，可参考下面一份配置
server: 作为 GRPC Server 独立启动时配置；托管给 baetyl-function-manager 无需配置
  address: GRPC Server 监听的地址，<host>:<port>
  workers:
    max: 默认 CPU 核数乘以5，线程池最大容量
  concurrent:
    max: 默认不限制，最大并发连接数
  message:
    length:
      max: 默认值：4m， 函数实例允许接收和发送的最大消息长度
  ca: Server 的 CA 证书路径
  key: Server 的服务端私钥路径
  cert: Server 的服务端公钥路径
functions: 函数列表
  - name: [必须] 函数名称，列表内唯一
    handler: [必须] 函数包和处理函数名，比如：'sayhi.handler'
    codedir: [必须] Python 代码所在路径
logger: 日志配置项
  path: 默认为空，即不打印到文件；如果指定文件则输出到文件
  level: 默认值：info，日志等级，支持 debug、info、warn 和 error
  format: 默认值：text，日志打印格式，支持 text 和 json
  age:
    max: 默认值：15，日志文件保留的最大天数
  size:
    max: 默认值：50，日志文件大小限制，单位MB
  backup:
        max: 默认值：15，日志文件保留的最大数量
```

## baetyl-remote-mqtt 配置

```yaml
hub:
  clientid: Client 连接 Hub 的 Client ID。cleansession 为 false 则不允许为空
  address: [必须] Client 连接 Hub 的地址
  username: [必须] Client 连接 Hub 的用户名
  password: 如果采用账号密码，必须填 Client 连接 Hub 的密码，否者不用填写
  ca: 如果采用证书双向认证，必须填 Client 连接 Hub 的 CA 证书路径
  key: 如果采用证书双向认证，必须填 Client 连接 Hub 的客户端私钥路径
  cert: 如果采用证书双向认证，必须填 Client 连接 Hub 的客户端公钥路径
  timeout: 默认值：30s，Client 连接 Hub 的超时时间
  interval: 默认值：1m，Client 连接 Hub 的重连最大间隔时间，从500微秒翻倍增加到最大值
  keepalive: 默认值：1m，Client 连接 Hub 的保持连接时间
  cleansession: 默认值：false，Client 连接 Hub 的是否保持 Session
  validatesubs: 默认值：false，Client 是否检查 Hub 订阅结果，如果是发现订阅失败报错退出
  buffersize: 默认值：10，Client 发送消息给 Hub 的内存队列大小，异常退出会导致消息丢失，恢复后 QoS 为1的消息依赖 Hub 重发
rules: 路由规则列表，向 Hub 订阅消息发送给 Remote，或反之
  - hub:
      clientid: Client 连接 Hub 的 Client ID
      subscriptions: Client 向 Hub 订阅的消息，例如
        - topic: say
          qos: 1
        - topic: hi
          qos: 0
    remote:
      name: [必须]指定Remote名称，必须是Remote列表中的一个
      clientid: Client 连接 Remote 的 Client ID
      subscriptions: Client 向 Remote 订阅的消息，例如
        - topic: remote/say
          qos: 0
        - topic: remote/hi
          qos: 0
remotes: Remote列表
  - name: [必须] Remote名称，列表内必须唯一
    clientid: Client 连接 Remote 的 Client ID
    address: [必须] Client 连接 Remote的地址
    username: Client 连接Remote的用户名
    password: 如果采用账号密码，必须填 Client 连接 Remote 的密码，否者不用填写
    ca: 如果采用证书双向认证，必须填 Client 连接 Remote 的 CA 证书路径
    key: 如果采用证书双向认证，必须填 Client 连接 Remote 的客户端私钥路径
    cert: 如果采用证书双向认证，必须填 Client 连接 Remote 的客户端公钥路径
    timeout: 默认值：30s，Client 连接 Remote 的超时时间
    interval: 默认值：1m，Client 连接 Remote 的重连最大间隔时间，从500微秒翻倍增加到最大值
    keepalive: 默认值：1m，Client 连接 Remote 的保持连接时间
    cleansession: 默认值：false，Client 连接 Remote 的是否保持 Session
    validatesubs: 默认值：false，Client 是否检查Remote订阅结果，如果是发现订阅失败报错退出
    buffersize: 默认值：10，Client 发送消息给 Remote 的内存队列大小，异常退出会导致消息丢失，恢复后 QoS 为1的消息依赖 Remote 重发
logger: 日志配置项
  path: 默认为空，即不打印到文件；如果指定文件则输出到文件
  level: 默认值：info，日志等级，支持 debug、info、warn 和 error
  format: 默认值：text，日志打印格式，支持 text 和 json
  age:
    max: 默认值：15，日志文件保留的最大天数
  size:
    max: 默认值：50，日志文件大小限制，单位 MB
  backup:
    max: 默认值：15，日志文件保留的最大数量
```

## baetyl-timer 配置

```yaml
hub: 要连接的 Hub 信息
  address: timer 模块要连接的 Hub 地址
  username: timer 模块接入 Hub 的用户名
  password: timer 模块接入 Hub 密码
  clientid: Client 连接 Hub 的 Client ID
timer: timer 相关属性
  interval: timer 模块定时间隔
publish:
  topic: 定时结果发布到 Hub 的主题
  payload: 负载数据，map 数据结构，例如：
    id: 1
logger: 日志配置项
  path: 默认为空，即不打印到文件；如果指定文件则输出到文件
  level: 默认值：info，日志等级，支持 debug、info、warn 和 error
```
