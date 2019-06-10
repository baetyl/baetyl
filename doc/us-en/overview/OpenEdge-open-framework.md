# OpenEdge Open Framework

OpenEdge provides an open framework, which allows access to any protocol through a variety of networks, and allows any application to run on multiple systems.

## OpenEdge Network Protocol Support

OpenEdge supports network protocols in 3 aspects: one is for IoT applications, OpenEdge provides device connect services based on Local Hub Module through MQTT protocol. And it supports `TCP`, `SSL`(TCP + SSL), `WS`(Websocket) and `WSS`(Websocket + SSL) 4 connection methods; the second is for the device hardware information reporting service, OpenEdge supports the HTTPS protocol; the other is for the remote service with the cloud, OpenEdge supports multiple network protocols. For example, OpenEdge publishes data message to the remote hub through Remote Module with MQTT protocol, publishes the data message to the remote Kafka through Kafka module, etc.

## OpenEdge System Platform Support

The biggest advantage of OpenEdge is that it can support seamlessly running on multiple operating systems and CPU platforms. The list of specific supported system platforms is as follows:

- Darwin-amd64
- Linux-386
- Linux-amd64
- Linux-armv7
- Linux-arm64

In particular, for Linux, OpenEdge only depends on the standard Linux kernel, and the version of Linux kernel should be higher than 2.6.32. In addition, for Docker container mode, OpenEdge also isolates and limits the resources of containers, and allocates the CPU, memory and other resources of each running instance accurately to improve the efficiency of resource utilization.
