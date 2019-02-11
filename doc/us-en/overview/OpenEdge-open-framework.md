# OpenEdge open framework

OpenEdge provides an open framework, which allows access to any protocol through a variety of networks, and allows any application to run on multiple systems.

## OpenEdge Network Protocol Support

OpenEdge supports network protocols in 3 aspects: one is for IoT applications, OpenEdge provides device connect services based on Local Hub module through MQTT protocol. And it supports `tcp`, `ssl`(tcp+ssl), `ws`(websocket) and `wss`(websocket+ssl) 4 connect methods; the second is for the device hardware information reporting service, OpenEdge supports the HTTPS protocol; the other is for the remote service with the cloud, OpenEdge supports multiple network protocols. For example, OpenEdge publishes data message to the remote hub through remote module with MQTT protocol, publishes the data message to the remote Kafka through Kafka module, etc.

## OpenEdge System Platform Support

The biggest advantage of OpenEdge is that it can support seamlessly running on multiple operating systems and CPU platforms. The list of specific supported system platforms is as follows:

> + Docker container mode
>   - Darwin-x86_64
>   - Linux-x86_64
>   - Linux-armv7
>   - Linux-aarch64
> + Native Process Mode
>   - Darwin-x86_64
>   - Linux-x86
>   - Linux-x86_64
>   - Linux-armv7
>   - Linux-aarch64
>   - Windows-x86
>   - Windows-x86_64

In particular, for Linux, OpenEdge only depends on the standard Linux kernel, and the version of Linux kernel should be higher than 2.6.32. In addition, for Docker container mode, OpenEdge also isolates and limits the resources of containers, and allocates the CPU, memory and other resources of each running instance accurately to improve the efficiency of resource utilization. **Note that**, Windows and Native process mode do not support resource isolation and restrictions.
