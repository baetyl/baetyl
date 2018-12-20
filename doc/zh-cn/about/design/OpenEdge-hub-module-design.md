# Hub模块（openedge_hub）

Hub模块是一个单机版的消息订阅和发布中心，采用MQTT 3.1.1协议，结构图如下：

![Hub模块结构图](../../images/about/hub.png)

目前支持4种接入方式：tcp、ssl(tcp+ssl)、ws(websocket)及wss(websocket+ssl)，MQTT协议支持度如下：

> - 支持connect、disconnect、subscribe、publish、unsubscribe、ping等功能
> - 支持QoS等级0和1的消息发布和订阅
> - 支持retain、will message、clean session
> - 支持订阅含有"+"、"#"等通配符的主题
> - 支持符合约定的clientid和payload的校验
> - 暂时**不支持**发布和订阅以"$"为前缀的主题
> - 暂时**不支持**client的keep alive特性以及QoS等级2的发布和订阅

_**注意**：_

> - 发布和订阅主题中含有的分隔符"/"最多不超过8个，主题名称长度最大不超过255个字符
> - 消息报文默认最大长度位32k，可支持的最大长度为268,435,455(byte)，不到256m，可通过message配置项进行修改
> - clientid支持大小写字母、数字、下划线、连字符（减号）和空字符(空字符表示client为临时连接，强制cleansession=true), 最大长度不超过128个字符
> - 消息的QoS只能降不能升，比如原消息的QoS为0时，即使订阅QoS为1，消息仍然以QoS为0的等级发送。

Hub模块支持简单的主题路由，比如订阅主题为t的消息并以新主题t/topic发布回broker。[参考配置](https://github.com/baidu/openedge/blob/master/example/docker/app/modu-nje2uoa9s/conf.yml)