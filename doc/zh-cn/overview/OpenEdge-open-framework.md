
# OpenEdge 开放式设计框架

OpenEdge 提供开放式的框架支持，允许通过各种网络类型接入任意协议，允许任意应用在任意系统平台上运行。

## OpenEdge 网络协议支持

OpenEdge 针对网络协议的支持，具体表现在三个方面：其一是针对物联网应用场景，OpenEdge 基于 Hub 模块提供设备接入服务，同时可支持 TCP、SSL（TCP + SSL）、WS(Websocket)及 WSS（Websocket + SSL）四种接入方式；其二是针对设备硬件信息上报服务，提供对 HTTPS 协议支持；另外就是针对与云端平台间的远程通讯服务，支持通过各种类型网络远程转发至云端平台，比如通过 MQTT 远程通讯模块将数据发送到云端 Hub，通过 Kafka 远程通讯模块将数据发送到云端 Kafka 等。

## OpenEdge 系统平台支持

OpenEdge 最大的特点和优势就是可以支持在多操作系统、多 CPU 平台上无缝运行。其具体可支持的系统平台列表如下：

> + Docker 容器模式
>   - Darwin-x86_64
>   - Linux-x86_64
>   - Linux-armv7
>   - Linux-aarch64
> + Native 进程模式
>   - Darwin-x86_64
>   - Linux-x86
>   - Linux-x86_64
>   - Linux-armv7
>   - Linux-aarch64
>   - Windows-x86
>   - Windows-x86_64

特别地，OpenEdge 对 Linux 各系统平台的适配支持，仅仅依赖于 Linux 标准内核，且版本高于 2.6.32 即可。此外，针对 Docker 容器模式，还支持资源隔离和限制（须启用 CGROUP，并以 root 权限启动），比如 CPU、内存等。**需要注意的是**，Windows 系统平台及 Native 进程模式暂不支持资源隔离和限制。