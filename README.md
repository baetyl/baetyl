# BAETYL v2

[![Baetyl-logo](./docs/logo_with_name.png)](https://baetyl.io)

[![build](https://github.com/baetyl/baetyl/workflows/build/badge.svg)](https://github.com/baetyl/baetyl/actions?query=workflow%3Abuild)
[![codecov](https://codecov.io/gh/baetyl/baetyl/branch/master/graph/badge.svg)](https://codecov.io/gh/baetyl/baetyl)
[![Go Report Card](https://goreportcard.com/badge/github.com/baetyl/baetyl)](https://goreportcard.com/report/github.com/baetyl/baetyl) 
[![License](https://img.shields.io/github/license/baetyl/baetyl?color=blue)](LICENSE) 
[![Stars](https://img.shields.io/github/stars/baetyl/baetyl?style=social)](Stars)

[![README_CN](https://img.shields.io/badge/README-%E4%B8%AD%E6%96%87-brightgreen)](./README_CN.md)

**[Baetyl](https://baetyl.io) is an open edge computing framework of
[Linux Foundation Edge](https://www.lfedge.org) that extends cloud computing,
data and service seamlessly to edge devices.** It can provide temporary offline, low-latency computing services include device connection, message routing, remote synchronization, function computing, video capture, AI inference, status reporting, configuration ota etc.

Baetyl v2 provides a new edge cloud integration platform, which adopts cloud management and edge operation solutions, and is divided into [**Edge Computing Framework (this project)**](https://github.com/baetyl/baetyl) and [**Cloud Management Suite**](https://github.com/baetyl/baetyl-cloud) supports varius deployment methods. It can manage all resources in the cloud, such as nodes, applications, configuration, etc., and automatically deploy applications to edge nodes to meet various edge computing scenarios. It is especially suitable for emerging strong edge devices, such as AI all-in-one machines and 5G roadside boxes.

The main differences between v2 and v1 versions are as follows:
* Edge and cloud frameworks have all evolved to cloud native, and already support running on K8S or K3S.
* Introduce declarative design, realize data synchronization (OTA) through shadow (Report/Desire).
* The edge framework currently supports Kube mode. Because it runs on K3S, the overall resource overhead is relatively large (1G memory); the Native mode is under development, which can greatly reduce resource consumption.
* The edge framework will support edge node clusters in the future.

## Architecture

![Architecture](./docs/baetyl-arch-v2.svg)

### [Edge Computing Framework (this project)](./README.md)

The Edge Computing Framework runs on Kubernetes at the edge node,
manages and deploys all applications which provide various capabilities.
Applications include system applications and common applications.
All system applications are officially provided by Baetyl,
and you do not need to configure them.

There are currently several system applications:
* baetyl-init: responsible for activating the edge node to the cloud
and initializing baetyl-core, and will exit after all tasks are completed.
* baetyl-core: responsible for local node management (node),
data synchronization with cloud (sync) and application deployment (engine).
* baetyl-function: the proxy for all function runtime services,
function invocations are passed through this module.

Currently the framework supports Linux/amd64, Linux/arm64, Linux/armv7,
If the resources of the edge nodes are limited,
consider to use the lightweight kubernetes: [K3S](https://k3s.io/).

Hardware requirements scale based on the size of your applications at edge. Minimum recommendations are outlined here.
* RAM: 1GB Minimum
* CPU: 1 Minimum

### [Cloud Management Suite](https://github.com/baetyl/baetyl-cloud)

The Cloud Management Suite is responsible for managing all resources, including nodes, applications, configuration, and deployment. The realization of all functions is plug-in, which is convenient for function expansion and third-party service access, and provides rich applications. The deployment of the cloud management suite is very flexible. It can be deployed on public clouds, private cloud environments, and common devices. It supports K8S/K3S deployment, and supports single-tenancy and multi-tenancy.

The basic functions provided by the cloud management suite in this project are as follows:
* Edge node management
     * Online installation of edge computing framework
     * Synchronization (shadow) between edge and cloud
     * Node information collection
     * Node status collection
     * Application status collection
* Application deployment management
     * Container application
     * Function application
     * Node matching (automatic)
* Configuration management
     * Common configuration
     * Function configuration
     * Secrets
     * Certificates
     * Registry credentials
* Node provisioning management
     * Node batch management
     * Registration and activation

_The open source version contains the RESTful API of all the above functions, but does not include the front-end dashboard. _

## Contact us

As the first open edge computing framework in China,
Baetyl aims to create a lightweight, secure,
reliable and scalable edge computing community
that will create a good ecological environment.
In order to create a better development of Baetyl,
if you have better advice about Baetyl, please contact us:

- Welcome to join [Baetyl's Wechat](https://baetyl.bj.bcebos.com/Wechat/Wechat-Baetyl.png)
- Welcome to join [Baetyl's LF Edge Community](https://lists.lfedge.org/g/baetyl/topics)
- Welcome to send email to <baetyl@lists.lfedge.org>
- Welcome to [submit an issue](https://github.com/baetyl/baetyl/issues)

## Contributing

If you are passionate about contributing to open source community,
Baetyl will provide you with both code contributions and document contributions.
More details, please see: [How to contribute code or document to Baetyl](./docs/contributing.md).