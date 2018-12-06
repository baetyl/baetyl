# 什么是OpenEdge

[OpenEdge](https://openedge.tech)是百度云发布的国内首个开源边缘计算产品，可将云计算能力拓展至用户现场，提供临时离线、低延时的计算服务，包括设备接入、消息路由、消息远程同步、函数计算等功能。OpenEdge和[智能边缘BIE](https://cloud.baidu.com/product/bie.html)（Baidu-IntelliEdge）云端管理套件配合使用，通过在云端进行智能边缘核心设备的建立、身份制定、策略规则制定、函数编写，然后生成配置文件下发至OpenEdge本地运行包，可达到云端管理和应用下发，边缘设备上运行应用的效果，满足各种边缘计算场景。

在架构设计上，OpenEdge一方面推行“模块化"，拆分各项主要功能，确保每一项功能都是一个独立的模块，整体由主程序模块控制启动、退出，确保各项子功能模块运行互不依赖、互不影响，总体上来说，推行模块化的设计模式，可以充分满足用户“按需使用、按需部署”的切实要求；另一方面，OpenEdge在设计上还采用“容器化"的设计思路，基于各模块提供的DockerFile文件可以在Docker支持的各类操作系统上进行“一键式部署”，依托Docker的跨平台支持特性，确保OpenEdge在各系统、各平台的环境一致性标准化；此外，OpenEdge还针对Docker容器化进行容器资源隔离与限制，精确分配各运行实例的CPU、内存等资源，提升资源利用效率。

## OpenEdge构成

OpenEdge主要由主程序模块、OpenEdge_Hub模块、OpenEdge_Function模块、OpenEdge_Remote_MQTT模块、OpenEdge_Function_Runtime_Python2.7模块构成。各模块的主要提供的能力如下：

- OpenEdge主程序模块负责所有模块的管理，如启动、退出等，由模块引擎、API构成；
	- 模块引擎负责模块的启动、停止、重启、监听和守护，目前支持Docker容器模式和Native进程模式；模块引擎从工作目录的配置文件中加载模块列表，并以列表的顺序逐个启动模块。模块引擎会为每个模块启动一个守护协程对模块状态进行监听，如果模块异常退出，会根据模块的Restart Policy配置项执行重启或退出。主程序关闭后模块引擎会按照列表的逆序逐个关闭模块；
	- OpenEdge主程序会暴露一组HTTP API，目前支持获取空闲端口，模块的启动、停止和重启。为了方便管理，我们对模块做了一个划分，从配置文件中加载的模块称为常驻模块，通过API启动的模块称为临时模块，临时模块遵循**“谁启动谁负责停止"**的原则。OpenEdge退出时，会先逆序停止所有常驻模块，常驻模块停止过程中也会调用API来停止其启动的模块，最后如果还有遗漏的临时模块，会随机全部停止。
- OpenEdge_Hub模块主要基于[MQTT协议](http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.html)提供设备接入（支持TCP、SSL（TCP+SSL）、WS（Websocket）及WSS（Websocket+SSL）四种接入方式）、消息路由转发等能力； 
- OpenEdge_Function提供基于MQTT消息机制，弹性、高可用、扩展性好、响应快的的计算能力，函数通过一个或多个具体的实例执行，每个实例都是一个独立的进程，现采用GRPC Server运行函数实例。所有函数实例由实例池（Pool）负责管理生命周期，支持自动扩容和缩容； 
- OpenEdge_Remote_MQTT模块目前支持MQTT协议，其实质是两个MQTT Server的桥接（Bridge）模块，用于订阅一个Server的消息并转发给另一个Server，特别地，； 
- OpenEdge_Function_Runtime_Python2.7是基于OpenEdge_Function模块的具体实例化表现形式，开发者通过编写的自己的函数来处理消息，可进行消息的过滤、转换和转发等，使用非常灵活。

## OpenEdge功能

 - **物联接入**：支持设备基于标准MQTT协议（V3.1和V3.1.1版本）与OpenEdge建立连接；
 - **消息转发**：通过消息路由转发机制，将数据转发至任意主题、计算函数；
 - **函数计算**：支持基于Python2.7及满足条件的任意自定义语言的函数编写、运行；
 - **远程同步**：支持与百度云天工IoTHub及符合OpenEdge_Remote_MQTT模块支持范围的远程消息同步。

## OpenEdge优势

 - **屏蔽计算框架**：OpenEdge提供主流运行时支持的同时，提供各类运行时转换服务，基于任意语言编写、基于任意框架训练的函数或模型，都可以在OpenEdge中执行；
 - **简化应用生产**：[智能边缘BIE](https://cloud.baidu.com/product/bie.html)云端管理套件配合OpenEdge，联合百度云，一起为OpenEdge提供强大的应用生产环境，通过[CFC](https://cloud.baidu.com/product/cfc.html)、[Infinite](https://cloud.baidu.com/product/infinite.html)、[Jarvis](http://di.baidu.com/product/jarvis)、[IoT EasyInsight](https://cloud.baidu.com/product/ist.html)、[TSDB](https://cloud.baidu.com/product/tsdb.html)、[IoT Visualization](https://cloud.baidu.com/product/iotviz.html)等产品，可以在云端轻松生产各类函数、AI模型，及将数据写入百度云天工云端TSDB及物可视进行展示；
 - **一键式运行环境部署**：OpenEdge推行Docker容器化，开发者可以根据OpenEdge源码包中各模块的DockerFile一键式构建OpenEdge运行环境；
 -  **按需部署**：OpenEdge推行功能模块化，各功能间运行互补影响、互不依赖，开发者完全可以根据自己的需求进行部署；
 - **丰富配置**：OpenEdge支持X86、ARM等多种硬件以及Linux、MacOS和Windows等主流操作系统。

## OpenEdge术语释义

- **OpenEdge**：百度开源边缘计算产品，提供设备接入、消息转发、计算等能力；
- **MQTT**：MQTT是基于二进制消息的发布/订阅（Publish/Subscribe）模式的协议，最早由IBM提出，如今已经成为OASIS规范，更符合M2M大规模沟通。