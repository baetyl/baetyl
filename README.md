# OpenEdge

[![OpenEdge Status](https://travis-ci.com/baidu/openedge.svg?branch=master)](https://travis-ci.com/baidu/openedge)

![OpenEdge-logo](./doc/images/logo/logo-with-name.png)

[README in Chinese](./README-CN.md)

**[OpenEdge](https://openedge.tech) is an open edge computing framework that extends cloud computing, data and service seamlessly to edge devices.** It can provide temporary offline, low-latency computing services, and include device connect, message routing, remote synchronization, function computing, video access pre-processing, AI inference, etc. The combination of OpenEdge and the **Cloud Management Suite** of [BIE](https://cloud.baidu.com/product/bie.html)(Baidu IntelliEdge) will achieve cloud management and application distribution, enable applications running on edge devices and meet all kinds of edge computing scenario.

## Features

> + support module management, include start, stop, restart, listen and upgrade
> + support two mode: **Docker Container Mode** and **Native Process Mode**
> + docker container mode support resources isolation and restriction
> + support cloud management suite, which can be used to report device hardware information and deploy configuration
> + provide **Local Hub Module**, which supports MQTT v3.1.1 protocol, qos 0 or 1, SSL/TLS authentication
> + provide **Local Function Module**, which supports function instance scaling, **Python2.7** runtime and customize runtime
> + provide **MQTT Remote Module**, which supports MQTT v3.1.1 protocol
> + provide **Module SDK(Golang)**, which can be used to develop customize module

## Advantages

> + **Shielding Computing Framework**: OpenEdge provides two official computing modules(**Local Function Module** and **Python Runtime Module**), also supports customize module(which can be written in any programming language or any machine learning framework).
> + **Simplified Application Production**: OpenEdge combines with **Cloud Management Suite** of BIE and many other productions of Baidu Cloud(such as [CFC](https://cloud.baidu.com/product/cfc.html), [Infinite](https://cloud.baidu.com/product/infinite.html), [Jarvis](http://di.baidu.com/product/jarvis), [IoT EasyInsight](https://cloud.baidu.com/product/ist.html), [TSDB](https://cloud.baidu.com/product/tsdb.html), [IoT Visualization](https://cloud.baidu.com/product/iotviz.html)) to provide data calculation, storage, visible display, model training and many more abilities.
> + **Quickly Deployment**: OpenEdge pursues docker container mode, it make developers quickly deploy OpenEdge on different operating system.
> + **Deploy On Demand**: OpenEdge takes modularization mode and splits functions to multiple independent modules. Developers can select some modules which they need to deploy.
> + **Rich Configuration**: OpenEdge supports X86 and ARM CPU processors, as well as Linux, Darwin and Windows operating systems.

## Install OpenEdge 

> + [Install OpenEdge on CentOS](./doc/us-en/setup/Install-OpenEdge-on-CentOS.md)
> + [Install OpenEdge on Debian](./doc/us-en/setup/Install-OpenEdge-on-Debian.md)
> + [Install OpenEdge on Raspbian](./doc/us-en/setup/Install-OpenEdge-on-Raspbian.md)
> + [Install OpenEdge on Ubuntu](./doc/us-en/setup/Install-OpenEdge-on-Ubuntu.md)
> + [Install OpenEdge on Darwin](./doc/us-en/setup/Install-OpenEdge-on-Darwin.md)
> + [Build OpenEdge from Source](./doc/us-en/setup/Build-OpenEdge-from-Source.md)

## Documents Of Design

> + [OpenEdge design](./doc/us-en/overview/OpenEdge-design.md)
> + [OpenEdge config interpretation](./doc/us-en/tutorials/Config-interpretation.md)
> + [How to write a python srcipt for python runtime](./doc/us-en/customize/How-to-write-a-python-script-for-python-runtime.md)
> + [How to develop a customize runtime for function](./doc/us-en/customize/How-to-develop-a-customize-runtime-for-function.md)
> + [How to develop a customize module for OpenEdge](./doc/us-en/customize/How-to-develop-a-customize-module-for-OpenEdge.md)


## Contributing

If you are passionate about contributing to open source community, OpenEdge will provide you with both code contributions and document contributions. More details, please see: [How to contribute code or document to OpenEdge](./doc/us-en/about/How-to-contribute.md).

## Copyright and License

OpenEdge is provided under the [Apache-2.0 license](./LICENSE).

## Discussion

As the first open edge computing framework in China, OpenEdge aims to create a lightweight, secure, reliable and scalable edge computing community that will create a good ecological environment. Here, we offer the following options for you to choose from:

> + If you want to participate in OpenEdge's daily development communication, you are welcome to join [Wechat-for-OpenEdge](https://openedge.bj.bcebos.com/Wechat/Wechat-OpenEdge.png)
> + If you have more about feature requirements or bug feedback of OpenEdge, please [Submit an issue](https://github.com/baidu/openedge/issues)
> + If you want to know more about OpenEdge and other services of Baidu Cloud, please visit [Baidu-Cloud-forum](https://cloud.baidu.com/forum/bce)
> + If you want to know more about Cloud Management Suite of BIE, please visit: [Baidu-IntelliEdge](https://cloud.baidu.com/product/bie.html)
> + If you have better development advice about OpenEdge, please contact us: [contact@openedge.tech](contact@openedge.tech)
