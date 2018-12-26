# 配置解读

- [说在前头](#说在前头)
- [主程序配置](#主程序配置)
- [应用配置](#应用配置)
- [Hub模块配置](#hub模块配置)
- [函数计算模块配置](#函数计算模块配置)
- [远程服务模块配置](#远程服务模块配置)
- [配置参考](#配置参考)

## 说在前头

- 配置中的路径可以是相对路径也可以是绝对路径，如果是相对路径则相对于工作目录而言的，容器模式下配置的路径表示容器中的路径。
- 支持的单位：
  - 大小：b：字节（byte）；k：千字节（kilobyte）；m：兆字节（megobyte），g：吉字节（gigabyte）
  - 时间：s：秒；m：分；h：小时

## 主程序配置

主程序的配置和应用配置是分离的，默认配置文件是：工作目录下的[etc/openedge/openedge.yml](https://github.com/baidu/openedge/blob/master/example/docker/openedge/openedge.yml)配置解读如下：

    mode: [必须]主程序运行模式。docker：docker容器模式；native：native进程模式
    grace: 默认值：30s，主程序平滑退出超时时间
    api: 主程序api server配置项
      address: 默认值可读取环境变量：OPENEDGE_MASTER_API，主程序api server地址。
      timeout: 默认值：30s，主程序api server超时时间
    cloud: 主程序对接云端管理套件的配置项，包括MQTT和HTTPS
      clientid: [必须]mqtt client连接云端的client id，必须是云端核心设备的ID
      address: [必须]mqtt client连接云端的接入地址，必须使用ssl endpoint
      username: [必须]mqtt client连接云端的用户名，必须是云端核心设备的用户名
      ca: [必须]mqtt client连接云端的CA证书所在路径
      key: [必须]mqtt client连接云端的客户端私钥所在路径
      cert: [必须]mqtt client连接云端的客户端公钥所在路径
      timeout: 默认值：30s，mqtt client连接云端的超时时间
      interval: 默认值：1m，mqtt client连接云端的重连最大间隔时间，从500微秒翻倍增加到最大值。
      keepalive: 默认值：1m，mqtt client连接云端的保持连接时间
      cleansession: 默认值：false，mqtt client连接云端的clean session
    logger: 日志配置项
      path: 默认为空，即不打印到文件；如果指定文件则输出到文件
      level: 默认值：info，日志等级，支持debug、info、warn和error
      format: 默认值：text，日志打印格式，支持text和json
      console: 默认值：false，日志是否输出到控制台
      age:
        max: 默认值：15，日志文件保留的最大天数
      size:
        max: 默认值：50，日志文件大小限制，单位MB
      backup:
        max: 默认值：15，日志文件保留的最大数量

## 应用配置

应用配置的默认配置文件是：工作目录下的[var/db/openedge/module/module.yml](https://github.com/baidu/openedge/blob/master/example/docker/var/db/openedge/module/module.yml)配置解读如下：

    version: 应用版本
    modules: 应用的模块列表
      - name: [必须]模块名，在模块列表中必须唯一
        entry: [必须]模块入口。docker容器模式下表示模块镜像；native进程模式下表示模块可执行程序所在路径
        restart: 模块的重启策略配置项
          retry:
            max: 默认为空，表示总是重试，模块重启最大次数
          policy: 默认值：always，重启策略。no：不重启；always：总是重启；on-failure：模块异常退出就重启；unless-stopped：模块正常退出就重启
          backoff:
            min: 默认值：1s，重启最小间隔时间
            max: 默认值：5m，重启最大间隔时间
            factor: 默认值：2，重启间隔增大倍数
        expose: Docker容器模式下暴露的端口，例如：
          - 0.0.0.0:1883:1883 # 如果不想暴露给宿主机外的设备访问，可以改成127.0.0.1:1883:1883
          - 0.0.0.0:1884:1884/tcp
          - 8080:8080/tcp
          - 8884:8884
        params: 模块的启动参数，例如：
          - '-c'
          - 'conf/conf.yml'
        env: 模块的环境变量，例如：
          version: v1
        resources: [只支持docker容器模式]模块的资源限制配置项
          cpu:
            cpus: 模块可用的CPU比例，例如：1.5，表示可以用1.5个CPU内核
            setcpus: 模块可用的CPU内核，例如：0-2，表示可以使用第0到2个CPU内核；0，表示可以使用第0个CPU内核；1，表示可以使用第1个CPU内核
          memory:
            limit: 模块可用的内存，例如：500m，表示可以用500兆内存
            swap: 模块可用的交换空间，例如：1g，表示可以用1G内存
          pids:
            limit: 模块可创建的进程数

## Hub模块配置

    name: [必须]模块名
    listen: [必须]监听地址，例如：
      - tcp://0.0.0.0:1883 # Native进程模式下，如果不想暴露给宿主机外的设备访问，可以改成tcp://127.0.0.1:1883
      - ssl://0.0.0.0:1884
      - ws://:8080/mqtt
      - wss://:8884/mqtt
    certificate: SSL/TLS证书认证配置项，如果启用ssl或wss必须配置
      ca: mqtt server CA证书所在路径
      key: mqtt server 服务端私钥所在路径
      cert: mqtt server 服务端公钥所在路径
    principals: 接入权限配置项，如果不配置则mqtt client无法接入hub，支持账号密码和证书认证
      - username: mqtt client接入hub的用户名
        password: mqtt client接入hub的密码
        permissions:
          - action: 操作权限。pub：发布权限；sub：订阅权限
            permit: 操作权限允许的主题列表，支持+和#匹配符
      - serialnumber: mqtt client使用证书双向认证接入hub时使用的client证书的serial number
        permissions:
          - action: 操作权限。pub：发布权限；sub：订阅权限
            permit: 操作权限允许的主题列表，支持+和#匹配符
    subscriptions: 主题路由配置项
      - source:
          topic: 订阅的主题
          qos: 订阅的QoS
        target:
          topic: 发布的主题
          qos: 发布的QoS
    message: 消息相关的配置项
      length:
        max: 默认值：32k；最大值：268,435,455字节(约256MB)，可允许传输的最大消息长度
      ingress: 消息接收配置项
        qos0:
          buffer:
            size: 默认值：10000，可缓存到内存中的QoS为0的消息数，增大缓存可提高消息接收的性能，若设备掉电，则会直接丢弃QoS为0的消息
        qos1:
          buffer:
            size:  默认值：100，等待持久化的QoS为1的消息缓存大小，增大缓存可提高消息接收的性能，但潜在的风险是模块异常退出（比如设备掉电）会丢失缓存的消息，不回复确认（puback）。模块正常退出会等待缓存的消息处理完，不会丢失数据。
          batch:
            max:  默认值：50，批量写QoS为1的消息到数据库（持久化）的最大条数，消息持久化成功后会回复确认（ack）
          cleanup:
            retention:  默认值：48h，QoS为1的消息保存在数据库中的时间，超过该时间的消息会在清理时物理删除
            interval:  默认值：1m，QoS为1的消息清理时间间隔
      egress: 消息发送配置项
        qos0:
          buffer:
            size:  默认值：10000，内存缓存中的待发送的QoS为0的消息数，若设备掉电会直接丢弃消息；缓存满后，新推送的消息直接丢弃
        qos1:
          buffer:
            size:  默认值：100，QoS为1的消息发送后，未确认（ack）的消息缓存大小，缓存满后，不再读取新消息，一直等待缓存中的消息被确认。QoS为1的消息发送给客户端成功后等待客户端确认（puback），如果客户端在规定时间内没有回复确认，i消息会一直重发，直到客户端回复确认或者session关闭
          batch:
            max:  默认值：50，批量从数据库读取消息的最大条数
          retry:
            interval:  默认值：20s，消息重发时间间隔
      offset: 消息序列号持久化相关配置
        buffer:
          size:  默认值：10000，被确认（ack）的消息的序列号的缓存队列大小。比如当前批量发送了QoS为1且序列号为1、2和3的三条消息给客户端，客户端确认了序列号1和3的消息，此时序列号1会入列并持久化，序列号3虽然已经确认，但是还是得等待序列号2被确认入列后才能入列。该设计可保证模块异常重启后仍能从持久化的序列号恢复消息处理，保证消息不丢，但是会出现消息重发，也因此暂不支持QoS为2的消息
        batch:
          max:  默认值：100，批量写消息序列号到数据库的最大条数
    logger: 日志配置项
      path: 默认为空，即不打印到文件；如果指定文件则输出到文件
      level: 默认值：info，日志等级，支持debug、info、warn和error
      format: 默认值：text，日志打印格式，支持text和json
      console: 默认值：false，日志是否输出到控制台
      age:
        max: 默认值：15，日志文件保留的最大天数
      size:
        max: 默认值：50，日志文件大小限制，单位MB
      backup:
        max: 默认值：15，日志文件保留的最大数量
    status: 模块状态配置项
      logging:
        enable: 默认值：false，是否打印iotedge状态信息
        interval: 默认值：60s，状态信息打印时间间隔
    storage: 数据库存储配置项
      dir: 默认值：var/db，数据库存储目录
    shutdown: 模块退出配置项
      timeout: 默认值：10m，模块退出超时时间

## 函数计算模块配置

    name: [必须]模块名
    hub:
      clientid: mqtt client连接hub的client id，如果为空则随机生成，且clean session强制变成true
      address: [必须]mqtt client连接hub的地址，docker容器模式下地址为hub模块名，native进程模式下为127.0.0.1
      username: 如果采用账号密码，必须填mqtt client连接hub的用户名
      password: 如果采用账号密码，必须填mqtt client连接hub的密码
      ca: 如果采用证书双向认证，必须填mqtt client连接hub的CA证书所在路径
      key: 如果采用证书双向认证，必须填mqtt client连接hub的客户端私钥所在路径
      cert: 如果采用证书双向认证，必须填mqtt client连接hub的客户端公钥所在路径
      timeout: 默认值：30s，mqtt client连接hub的超时时间
      interval: 默认值：1m，mqtt client连接hub的重连最大间隔时间，从500微秒翻倍增加到最大值。
      keepalive: 默认值：1m，mqtt client连接hub的保持连接时间
      cleansession: 默认值：false，mqtt client连接hub的clean session
      buffersize: 默认值：10，mqtt client发送消息给hub的内存队列大小，异常退出会导致消息丢失，恢复后QoS为1的消息依赖hub重发
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
        runtime: 配置函数依赖的runtime模块名称，sql即为'sql'，tensorflow为'tensorflow'，python则为'python2.7'
        entry: 同模块的entry，运行函数实例的runtime模块的镜像或可执行程序
        handler: [必须]函数处理函数。sql为sql语句，比如：'select uuid() as id, topic() as topic, * where id < 10'；python为函数包和处理函数名，比如：'sayhi.handler'；tensorflow对应的配置信息为'tag:input_tensor:output_tensor'，其中tag为模型标签，input_tensor为模型网络结构的输入节点Tensor名称，output_tensor为模型网络结构的输出节点Tensor名称，强制模型"单一输入、单一输出"
        codedir: 如果是python，必须填python代码所在路径；如果是tensorflow，必须填写AI推断模型所在路径，且要求模型文件名必须为saved_model.pb
        env: 环境变量配置项，例如：
          USER_ID: acuiot
        instance: 函数实例配置项
          min: 默认值：0，最小值：0，最大值：100，最小函数实例数
          max: 默认值：1，最小值：1，最大值：100，最大函数实例数
          timeout: 默认值：5m， 函数实例调用超时时间
          idletime: 默认值：10m，函数实例最大空闲时间，超过后销毁，定期检查的时间间隔是idletime的一半。
          message:
            length:
              max: 默认值：4m， 函数实例允许接收和发送的最大消息长度
          cpu: [只支持docker容器模式]
            cpus: 函数实例模块可用的CPU比例，例如：1.5，表示可以用1.5个CPU内核
            setcpus: 函数实例模块可用的CPU内核，例如：0-2，表示可以使用第0到2个CPU内核；0，表示可以使用第0个CPU内核；1，表示可以使用第1个CPU内核
          memory: [只支持docker容器模式]
            limit: 函数实例模块可用的内存，例如：500m，表示可以用500兆内存
            swap: 函数实例模块可用的交换空间，例如：1g，表示可以用1G内存
          pids: [只支持docker容器模式]
            limit: 函数实例模块可创建的进程数
    logger: 日志配置项
      path: 默认为空，即不打印到文件；如果指定文件则输出到文件
      level: 默认值：info，日志等级，支持debug、info、warn和error
      format: 默认值：text，日志打印格式，支持text和json
      console: 默认值：false，日志是否输出到控制台
      age:
        max: 默认值：15，日志文件保留的最大天数
      size:
        max: 默认值：50，日志文件大小限制，单位MB
      backup:
        max: 默认值：15，日志文件保留的最大数量

## 远程服务模块配置

    name: [必须]模块名
    hub:
      clientid: mqtt client连接hub的client id，如果为空则随机生成，且clean session强制变成true
      address: [必须]mqtt client连接hub的地址，docker容器模式下地址为hub模块名，native进程模式下为127.0.0.1
      username: 如果采用账号密码，必须填mqtt client连接hub的用户名
      password: 如果采用账号密码，必须填mqtt client连接hub的密码
      ca: 如果采用证书双向认证，必须填mqtt client连接hub的CA证书所在路径
      key: 如果采用证书双向认证，必须填mqtt client连接hub的客户端私钥所在路径
      cert: 如果采用证书双向认证，必须填mqtt client连接hub的客户端公钥所在路径
      timeout: 默认值：30s，mqtt client连接hub的超时时间
      interval: 默认值：1m，mqtt client连接hub的重连最大间隔时间，从500微秒翻倍增加到最大值。
      keepalive: 默认值：1m，mqtt client连接hub的保持连接时间
      cleansession: 默认值：false，mqtt client连接hub的clean session
      buffersize: 默认值：10，mqtt client发送消息给hub的内存队列大小，异常退出会导致消息丢失，恢复后QoS为1的消息依赖remote重发
      subscriptions: 订阅配置项
        - topic: 向hub订阅消息的主题
          qos: 向hub订阅消息的QoS
    remote:
      clientid: mqtt client连接remote的client id，如果为空则随机生成，且clean session强制变成true
      address: [必须]mqtt client连接remote的地址
      username: 如果采用账号密码，必须填mqtt client连接remote的用户名
      password: 如果采用账号密码，必须填mqtt client连接remote的密码
      ca: 如果采用证书双向认证，必须填mqtt client连接remote的CA证书所在路径
      key: 如果采用证书双向认证，必须填mqtt client连接remote的客户端私钥所在路径
      cert: 如果采用证书双向认证，必须填mqtt client连接remote的客户端公钥所在路径
      timeout: 默认值：30s，mqtt client连接remote的超时时间
      interval: 默认值：1m，mqtt client连接remote的重连最大间隔时间，从500微秒翻倍增加到最大值。
      keepalive: 默认值：1m，mqtt client连接remote的保持连接时间
      cleansession: 默认值：false，mqtt client连接remote的clean session
      buffersize: 默认值：10，mqtt client发送消息给remote的内存队列大小，异常退出会导致消息丢失，恢复后QoS为1的消息依赖hub重发
      subscriptions: 订阅配置项
        - topic: 向remote订阅消息的主题
          qos: 向remote订阅消息的QoS
    logger: 日志配置项
      path: 默认为空，即不打印到文件；如果指定文件则输出到文件
      level: 默认值：info，日志等级，支持debug、info、warn和error
      format: 默认值：text，日志打印格式，支持text和json
      console: 默认值：false，日志是否输出到控制台
      age:
        max: 默认值：15，日志文件保留的最大天数
      size:
        max: 默认值：50，日志文件大小限制，单位MB
      backup:
        max: 默认值：15，日志文件保留的最大数量

## 配置参考

> - [Docker容器模式配置举例](https://github.com/baidu/openedge/blob/master/example/docker/etc/openedge/openedge.yml)
> - [Native容器模式配置举例](https://github.com/baidu/openedge/blob/master/example/native/etc/openedge/openedge.yml)
