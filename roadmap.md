# Roadmap

## 2019 Goals

Baetyl will embrace cloud native to edge creative, including:

- Deliver a lightweight, immutable, all-in-memory edge computing operating system based on Linux kernel.
- Build and manage service mesh across cloud, edge and IoT devices to enable free flow of applications and data.
- Improved troubleshooting and tunability of Baetyl under production conditions.
- Supports edge clusters consisting of a large number of nodes, providing intelligent task scheduling capabilities.
- Provides various forms of device and data connection capabilities such as AMQP, OPC-UA, and ZigBee.

## Release 1.0 in Autumn 2019

Baetyl will provide infrastructure and application development tools for IoT and edge computing, including:

- Core framework:
  - System security
    - Support for edge device activation and dynamic delivery of unique certificates
    - Provides a mechanism for bulk device activation
  - Application support
    - Logs can be viewed remotely, support for application acquisition, and can be set to cut off according to factors such as capacity and time
    - Provide inter-service dependencies and support the use of ports for health checks
    - Ability to pre-download application images
  - Remote management
    - Provide open source remote management console
- System related capabilities
  - Linux platform
    - Migrate to containerd, eliminate dependence on docker, further lightweight
    - Elimination of transport overhead introduced by virtual machines in container mode
    - Support for native package management on major Linux distributions
    - Support x86/x86-64/armv7/aarch64/mips/mipsle/mips64/mips64le
  - Windows platform
    - Support WSL+Docker
    - Experimental support for Windows Container
    - Experimental support for Windows 10 IoT Core
  - Kubernetes platform
    - Experimental support for K3S
    - Experimental support function distributed on multiple nodes
- Application development support
  - SDK
    - SDK with C interface
    - API for calling function in function
    - API for invoking ML inference
    - API for image capture and processing
    - API for local KV storage
  - South messaging service
    - Further reduce performance overhead on resource-limited devices
  - Function calculation
    - Support for binary functions such as C++
    - Support for shell functions
  - Streaming calculation
    - Support Flink SQL compatible streaming computing
  - North data synchronization
    - Support for connection to Mosquito, EMQ and AWS's IOTHUB
    - Support for connection to Kafka
