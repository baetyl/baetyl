
# OpenEdge开放式设计框架

OpenEdge提供开放式的框架支持，允许通过各种网络类型接入任意协议，允许任意应用在任意系统平台上运行。

## OpenEdge网络协议支持

OpenEdge针对网络协议的支持，具体表现在三个方面：其一是针对物联网应用场景，OpenEdge基于Hub模块提供设备接入服务，同时可支持tcp、ssl(tcp+ssl)、ws(websocket)及wss(websocket+ssl)四种接入方式；其二是针对设备硬件信息上报服务，提供对HTTPS协议支持；另外就是针对与云端平台间的远程通讯服务，支持通过各种类型网络远程转发至云端平台，比如通过MQTT远程通讯模块将数据发送到云端Hub，通过Kafka远程通讯模块将数据发送到云端Kafka，通过TSDB远程通讯模块将数据发送到云端TSDB等。另外还可以通过实现写云端服务的函数来实现数据的传输。

## OpenEdge系统平台支持

OpenEdge最大的特点和优势就是可以支持在多操作系统、多CPU平台上无缝运行。其具体可支持的系统平台列表如下：

> + Docker容器模式
>   - Darwin-x86_64
>   - Linux-x86_64
>   - Linux-armv7
>   - Linux-aarch64
> + Native进程模式
>   - Darwin-x86_64
>   - Linux-x86
>   - Linux-x86_64
>   - Linux-armv7
>   - Linux-aarch64
>   - Windows-x86
>   - Windows-x86_64

特别地，OpenEdge对Linux各系统平台的适配支持，仅仅依赖于Linux标准内核，且版本高于2.6.32即可。此外，针对Docker容器模式，还支持资源隔离和限制（须启用CGROUP，并以root权限启动），比如CPU、内存等。**需要注意的是**，Windows系统平台及Native进程模式暂不支持资源隔离和限制。