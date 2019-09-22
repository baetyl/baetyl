# Baetyl

![Travis (.org) branch](https://img.shields.io/travis/baetyl/baetyl/master) [![Documentation Status](https://img.shields.io/badge/docs-latest-brightgreen.svg?style=flat)](https://docs.baetyl.io/en/latest/) [![Release](https://img.shields.io/github/v/release/baetyl/baetyl?color=blue&include_prereleases&label=pre-release)](https://github.com/baetyl/baetyl/releases) [![License](https://img.shields.io/github/license/baetyl/baetyl?color=blue)](LICENSE) [![Stars](https://img.shields.io/github/stars/baetyl/baetyl?style=social)](Stars)

![Baetyl-logo](./logo-with-name.png)

[README 中文](./README-CN.md)

**[Baetyl](https://baetyl.io) is an open edge computing framework of [Linux Foundation Edge](https://www.lfedge.org) that extends cloud computing, data and service seamlessly to edge devices.** It can provide temporary offline, low-latency computing services, and include device connect, message routing, remote synchronization, function computing, video access pre-processing, AI inference, device resources report etc. The combination of Baetyl and the **Cloud Management Suite** of [BIE](https://cloud.baidu.com/product/bie.html)(Baidu IntelliEdge) will achieve cloud management and application distribution, enable applications running on edge devices and meet all kinds of edge computing scenario.

About architecture design, Baetyl takes **modularization** and **containerization** design mode. Based on the modular design pattern, Baetyl splits the product to multiple modules, and make sure each one of them is a separate, independent module. In general, Baetyl can fully meet the conscientious needs of users to deploy on demand. Besides, Baetyl also takes containerization design mode to build images. Due to the cross-platform characteristics of docker to ensure the running environment of each operating system is consistent. In addition, **Baetyl also isolates and limits the resources of containers**, and allocates the CPU, memory and other resources of each running instance accurately to improve the efficiency of resource utilization.

## Advantages

- **Shielding Computing Framework**: Baetyl provides two official computing modules(**Local Function Module** and **Python Runtime Module**), also supports customize module(which can be written in any programming language or any machine learning framework).
- **Simplify Application Production**: Baetyl combines with **Cloud Management Suite** of BIE and many other productions of Baidu Cloud(such as [CFC](https://cloud.baidu.com/product/cfc.html), [Infinite](https://cloud.baidu.com/product/infinite.html), [EasyEdge](https://ai.baidu.com/easyedge/home), [TSDB](https://cloud.baidu.com/product/tsdb.html), [IoT Visualization](https://cloud.baidu.com/product/iotviz.html)) to provide data calculation, storage, visible display, model training and many more abilities.
- **Service Deployment on Demand**: Baetyl adopts containerization and modularization design, and each module runs independently and isolated. Developers can choose modules to deploy based on their own needs.
- **Support multiple platforms**: Baetyl supports multiple hardware and software platforms, such as X86 and ARM CPU, Linux and Darwin operating systems.

## Contributing

If you are passionate about contributing to open source community, Baetyl will provide you with both code contributions and document contributions. More details, please see: [How to contribute code or document to Baetyl](./Contributing.md).

## Contact us

As the first open edge computing framework in China, Baetyl aims to create a lightweight, secure, reliable and scalable edge computing community that will create a good ecological environment. In order to create a better development of Baetyl, if you have better advice about Baetyl, please contact us:

- Welcome to join [Baetyl's Wechat](https://baetyl.bj.bcebos.com/Wechat/Wechat-Baetyl.png)
- Welcome to join [Baetyl's LF Edge Community](https://lists.lfedge.org/g/baetyl/topics)
- Welcome to send email to <baetyl@lists.lfedge.org>
- Welcome to [submit an issue](https://github.com/baetyl/baetyl/issues)

