# OpenEdge 构成

OpenEdge主要由主程序模块和若干功能模块构成，目前官方提供本地Hub、本地函数计算（和多种函数计算运行时）、远程通讯模块等。各模块的主要提供的能力如下：

- OpenEdge主程序模块负责所有模块的管理，如启动、退出等，由模块引擎、API、云代理构成；
	- [模块引擎](./OpenEdge-design.md#模块引擎(engine))负责模块的启动、停止、重启、监听和守护，目前支持Docker容器模式和Native进程模式；模块引擎从工作目录的配置文件中加载模块列表，并以列表的顺序逐个启动模块。模块引擎会为每个模块启动一个守护协程对模块状态进行监听，如果模块异常退出，会根据模块的 [Restart Policy](https://github.com/baidu/openedge/blob/master/module/config/policy.go) 配置项执行重启或退出。主程序关闭后模块引擎会按照列表的逆序逐个关闭模块；
	- [云代理](./OpenEdge-design.md#云代理(agent))负责和云端管理套件通讯，走MQTT和HTTPS通道，MQTT强制SSL/TLS证书双向认证，HTTPS强制SSL/TLS证书单向认证。OpenEdge启动和热加载（reload）完成后会通过云代理上报一次设备信息；
	- OpenEdge主程序会暴露一组 [HTTP API](./OpenEdge-design.md#API(api))，目前支持获取空闲端口，模块的启动和停止。为了方便管理，我们对模块做了一个划分，从配置文件中加载的模块称为常驻模块，通过API启动的模块称为临时模块，临时模块遵循“**谁启动谁负责停止**"的原则。OpenEdge退出时，会先逆序停止所有常驻模块，常驻模块停止过程中也会调用API来停止其启动的模块，最后如果还有遗漏的临时模块，会随机全部停止。
- 本地 Hub 模块提供基于 [MQTT 协议](http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.html) 的订阅和发布功能，支持TCP、SSL（TCP+SSL）、WS（Websocket）及WSS（Websocket+SSL）四种接入方式、消息路由转发等功能；
- 本地函数计算模块提供基于 MQTT 消息机制，弹性、高可用、扩展性好、响应快的的计算能力，函数通过一个或多个具体的实例执行，每个实例都是一个独立的进程，现采用 GRPC Server 运行函数实例。所有函数实例由实例池（Pool）负责管理生命周期，支持自动扩容和缩容；
- 远程通讯模块目前支持 MQTT 协议，其实质是两个 MQTT Server 的桥接（Bridge）模块，用于订阅一个 Server 的消息并转发给另一个 Server；目前支持配置多路消息转发，可配置多个 Remote 和 Hub 同时进行消息同步；
- 函数计算 Python 运行时模块是本地函数计算模块的具体实例化表现形式，开发者通过编写的自己的 Python 函数来处理消息，可进行消息的过滤、转换和转发等，使用非常灵活。