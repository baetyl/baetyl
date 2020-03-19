# Baetyl V1

[![Baetyl-logo](./docs/images/logo/logo-with-name.png)](https://baetyl.io)

[![build](https://github.com/baetyl/baetyl/workflows/build/badge.svg)](https://github.com/baetyl/baetyl/actions?query=workflow%3Abuild)
[![codecov](https://codecov.io/gh/baetyl/baetyl/branch/master/graph/badge.svg)](https://codecov.io/gh/baetyl/baetyl)
[![Go Report Card](https://goreportcard.com/badge/github.com/baetyl/baetyl)](https://goreportcard.com/report/github.com/baetyl/baetyl) 
[![Release](https://img.shields.io/github/v/release/baetyl/baetyl?color=blue&label=release)](https://github.com/baetyl/baetyl/releases) 
[![License](https://img.shields.io/github/license/baetyl/baetyl?color=blue)](LICENSE) 
[![Stars](https://img.shields.io/github/stars/baetyl/baetyl?style=social)](Stars)

[![Documentation in English](https://img.shields.io/badge/docs%20in%20English-latest-brightgreen)](https://docs.baetyl.io/en/latest/) [![中文文档](https://img.shields.io/badge/%E4%B8%AD%E6%96%87%E6%96%87%E6%A1%A3-%E6%9C%80%E6%96%B0-brightgreen)](https://docs.baetyl.io/zh_CN/latest/)


**[Baetyl](https://baetyl.io) is an open edge computing framework of [Linux Foundation Edge](https://www.lfedge.org) that extends cloud computing, data and service seamlessly to edge devices.** It can provide temporary offline, low-latency computing services, and include device connect, message routing, remote synchronization, function computing, video access pre-processing, AI inference, device resources report etc. The combination of Baetyl and the **Cloud Management Suite** of [BIE](https://cloud.baidu.com/product/bie.html)(Baidu IntelliEdge) will achieve cloud management and application distribution, enable applications running on edge devices and meet all kinds of edge computing scenario.

About architecture [design](./docs/overview/Design.md), Baetyl takes **modularization** and **containerization** design mode. Based on the modular design pattern, Baetyl splits the product to multiple modules, and make sure each one of them is a separate, independent module. In general, Baetyl can fully meet the conscientious needs of users to deploy on demand. Besides, Baetyl also takes containerization design mode to build images. Due to the cross-platform characteristics of docker to ensure the running environment of each operating system is consistent. In addition, **Baetyl also isolates and limits the resources of containers**, and allocates the CPU, memory and other resources of each running instance accurately to improve the efficiency of resource utilization.

## Advantages

- **Shielding Computing Framework**: Baetyl provides two official computing modules(**Local Function Module** and **Python Runtime Module**), also supports customize module(which can be written in any programming language or any machine learning framework).
- **Simplify Application Production**: Baetyl combines with **Cloud Management Suite** of BIE and many other productions of Baidu Cloud(such as [CFC](https://cloud.baidu.com/product/cfc.html), [Infinite](https://cloud.baidu.com/product/infinite.html), [EasyEdge](https://ai.baidu.com/easyedge/home), [TSDB](https://cloud.baidu.com/product/tsdb.html), [IoT Visualization](https://cloud.baidu.com/product/iotviz.html)) to provide data calculation, storage, visible display, model training and many more abilities.
- **Service Deployment on Demand**: Baetyl adopts containerization and modularization design, and each module runs independently and isolated. Developers can choose modules to deploy based on their own needs.
- **Support multiple platforms**: Baetyl supports multiple hardware and software platforms, such as X86 and ARM CPU, Linux and Darwin operating systems.

## Components

As an edge computing platform, **Baetyl** not only provides features such as underlying service management, but also provides some basic functional modules, as follows:

- Baetyl [Master](./docs/overview/Design.md#master) is responsible for the management of service instances, such as start, stop, supervise, etc., consisting of Engine, API, Command Line. And supports two modes of running service: **native** process mode and **docker** container mode
- The official module [baetyl-agent](./docs/overview/Design.md#baetyl-agent) is responsible for communication with the BIE cloud management suite, which can be used for application delivery, device information reporting, etc. Mandatory certificate authentication to ensure transmission security;
- The official module [baetyl-hub](./docs/overview/Design.md#baetyl-hub) provides message subscription and publishing functions based on the [MQTT protocol](http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.html), and supports four access methods: TCP, SSL, WS, and WSS;
- The official module [baetyl-remote-mqtt](./docs/overview/Design.md#baetyl-remote-mqtt) is used to bridge two MQTT Servers for message synchronization and supports configuration of multiple message route rules. ;
- The official module [baetyl-function-manager](./docs/overview/Design.md#baetyl-function-manager) provides computing power based on MQTT message mechanism, flexible, high availability, good scalability, and fast response;
- The official module [baetyl-function-python27](./docs/overview/Design.md#baetyl-function-python27) provides the Python2.7 function runtime, which can be dynamically started by `baetyl-function-manager`;
- The official module [baetyl-function-python36](./docs/overview/Design.md#baetyl-function-python36) provides the Python3.6 function runtime, which can be dynamically started by `baetyl-function-manager`;
- The official module [baetyl-function-node85](./docs/overview/Design.md#baetyl-function-node85) provides the Node 8.5 function runtime, which can be dynamically started by `baetyl-function-manager`;
- SDK (Golang) can be used to develop custom modules.

### Architecture

![Architecture](./docs/images/overview/design/design_overview.png)

## Installation

- [Quick Install Baetyl](https://docs.baetyl.io/en/latest/install/Quick-Install.html)
- [Install Baetyl From source](https://docs.baetyl.io/en/latest/install/Install-from-source.html)


## Guides

- [Baetyl configuration interpretation](https://docs.baetyl.io/en/latest/guides/Config-interpretation.html)
- [Device connect to Hub Service](https://docs.baetyl.io/en/latest/guides/Device-connect-to-hub-service.html)
- [Message transferring among devices with Hub Service](https://docs.baetyl.io/en/latest/guides/Message-transfer-among-devices-with-hub-service.html)
- [Message handling with Function Service](https://docs.baetyl.io/en/latest/guides/Message-handling-with-function-service.html)
- [Message Synchronize between baetyl-hub and Baidu IoTHub via Remote Service](https://docs.baetyl.io/en/latest/guides/Message-synchronize-with-iothub-through-remote-service.html)
- [Image capturing and AI model inference with Video infer Service](https://docs.baetyl.io/en/latest/guides/Image-capturing-and-AI-model-inference-with-video-infer-service.html)

## Development

- [Baetyl design](./docs/overview/Design.md)
- [How to write Python script for Python runtime](https://docs.baetyl.io/en/latest/develop/How-to-write-a-python-script-for-python-runtime.html)
- [How to write Node script for Node runtime](https://docs.baetyl.io/en/latest/develop/How-to-write-a-node-script-for-node-runtime.html)
- [How to import third-party libraries for Python runtime](https://docs.baetyl.io/en/latest/develop/How-to-import-third-party-libraries-for-python-runtime.md)
- [How to import third-party libraries for Node runtime](https://docs.baetyl.io/en/latest/develop/How-to-import-third-party-libraries-for-node-runtime.md)
- [How to develop a customize runtime for function](https://docs.baetyl.io/en/latest/develop/How-to-develop-a-customize-runtime-for-function.md)
- [How to develop a customize module for Baetyl](https://docs.baetyl.io/en/latest/develop/How-to-develop-a-customize-module.md)

## Contributing

If you are passionate about contributing to open source community, Baetyl will provide you with both code contributions and document contributions. More details, please see: [How to contribute code or document to Baetyl](./docs/overview/Contributing.md).

## Contact us

As the first open edge computing framework in China, Baetyl aims to create a lightweight, secure, reliable and scalable edge computing community that will create a good ecological environment. In order to create a better development of Baetyl, if you have better advice about Baetyl, please contact us:

- Welcome to join [Baetyl's Wechat](https://baetyl.bj.bcebos.com/Wechat/Wechat-Baetyl.png)
- Welcome to join [Baetyl's LF Edge Community](https://lists.lfedge.org/g/baetyl/topics)
- Welcome to send email to <baetyl@lists.lfedge.org>
- Welcome to [submit an issue](https://github.com/baetyl/baetyl/issues)

