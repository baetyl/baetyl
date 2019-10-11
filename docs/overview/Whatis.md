# What is Baetyl

**[Baetyl](https://baetyl.io) is an open edge computing framework of [Linux Foundation Edge](https://www.lfedge.org) that extends cloud computing, data and service seamlessly to edge devices.** It can provide temporary offline, low-latency computing services, and include device connect, message routing, remote synchronization, function computing, video access pre-processing, AI inference, device resources report etc. The combination of Baetyl and the **Cloud Management Suite** of [BIE](https://cloud.baidu.com/product/bie.html)(Baidu IntelliEdge) will achieve cloud management and application distribution, enable applications running on edge devices and meet all kinds of edge computing scenario.

About architecture [design](Design.md), Baetyl takes **modularization** and **containerization** design mode. Based on the modular design pattern, Baetyl splits the product to multiple modules, and make sure each one of them is a separate, independent module. In general, Baetyl can fully meet the conscientious needs of users to deploy on demand. Besides, Baetyl also takes containerization design mode to build images. Due to the cross-platform characteristics of docker to ensure the running environment of each operating system is consistent. In addition, **Baetyl also isolates and limits the resources of containers**, and allocates the CPU, memory and other resources of each running instance accurately to improve the efficiency of resource utilization.

## Advantages

- **Shielding Computing Framework**: Baetyl provides two official computing modules(**Local Function Module** and **Python Runtime Module**), also supports customize module(which can be written in any programming language or any machine learning framework).
- **Simplify Application Production**: Baetyl combines with **Cloud Management Suite** of BIE and many other productions of Baidu Cloud(such as [CFC](https://cloud.baidu.com/product/cfc.html), [Infinite](https://cloud.baidu.com/product/infinite.html), [EasyEdge](https://ai.baidu.com/easyedge/home), [TSDB](https://cloud.baidu.com/product/tsdb.html), [IoT Visualization](https://cloud.baidu.com/product/iotviz.html)) to provide data calculation, storage, visible display, model training and many more abilities.
- **Service Deployment on Demand**: Baetyl adopts containerization and modularization design, and each module runs independently and isolated. Developers can choose modules to deploy based on their own needs.
- **Support multiple platforms**: Baetyl supports multiple hardware and software platforms, such as X86 and ARM CPU, Linux and Darwin operating systems.

## Components

As an edge computing platform, **Baetyl** not only provides features such as underlying service management, but also provides some basic functional modules, as follows:

- Baetyl [Master](Design.html#master) is responsible for the management of service instances, such as start, stop, supervise, etc., consisting of Engine, API, Command Line. And supports two modes of running service: **native** process mode and **docker** container mode
- The official module [baetyl-agent](Design.html#baetyl-agent) is responsible for communication with the BIE cloud management suite, which can be used for application delivery, device information reporting, etc. Mandatory certificate authentication to ensure transmission security;
- The official module [baetyl-hub](Design.html#baetyl-hub) provides message subscription and publishing functions based on the [MQTT protocol](http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.html), and supports four access methods: TCP, SSL, WS, and WSS;
- The official module [baetyl-remote-mqtt](Design.html#baetyl-remote-mqtt) is used to bridge two MQTT Servers for message synchronization and supports configuration of multiple message route rules. ;
- The official module [baetyl-function-manager](Design.html#baetyl-function-manager) provides computing power based on MQTT message mechanism, flexible, high availability, good scalability, and fast response;
- The official module [baetyl-function-python27](Design.html#baetyl-function-python27) provides the Python2.7 function runtime, which can be dynamically started by `baetyl-function-manager`;
- The official module [baetyl-function-python36](Design.html#baetyl-function-python36) provides the Python3.6 function runtime, which can be dynamically started by `baetyl-function-manager`;
- The official module [baetyl-function-node85](Design.html#baetyl-function-node85) provides the Node 8.5 function runtime, which can be dynamically started by `baetyl-function-manager`;
- SDK (Golang) can be used to develop custom modules.

### Architecture

![Architecture](../images/overview/design/design_overview.png)

## Installation

- [Quick Install Baetyl](../install/Quick-Install.md)
- [Install Baetyl from source](../install/Install-from-source.md)

## Development

- [Baetyl design](Design.md)
- [Baetyl config interpretation](../guides/Config-interpretation.md)
- [How to write Python script for Python runtime](../develop/How-to-write-a-python-script-for-python-runtime.md)
- [How to write Node script for Node runtime](../develop/How-to-write-a-node-script-for-node-runtime.md)
- [How to import third-party libraries for Python runtime](../develop/How-to-import-third-party-libraries-for-python-runtime.md)
- [How to import third-party libraries for Node runtime](../develop/How-to-import-third-party-libraries-for-node-runtime.md)
- [How to develop a customize runtime for function](../develop/How-to-develop-a-customize-runtime-for-function.md)
- [How to develop a customize module for Baetyl](../develop/How-to-develop-a-customize-module.md)

## Contributing

If you are passionate about contributing to open source community, Baetyl will provide you with both code contributions and document contributions. More details, please see: [How to contribute code or document to Baetyl](Contributing.md).

## Contact us

As the first open edge computing framework in China, Baetyl aims to create a lightweight, secure, reliable and scalable edge computing community that will create a good ecological environment. In order to create a better development of Baetyl, if you have better advice about Baetyl, please contact us:

- Welcome to join [Baetyl's Wechat](https://baetyl.bj.bcebos.com/Wechat/Wechat-Baetyl.png)
- Welcome to join [Baetyl's LF Edge Community](https://lists.lfedge.org/g/baetyl/topics)
- Welcome to send email to <baetyl@lists.lfedge.org>
- Welcome to [submit an issue](https://github.com/baetyl/baetyl/issues)


