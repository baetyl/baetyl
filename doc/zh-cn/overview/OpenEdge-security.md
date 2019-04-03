
# OpenEdge 安全

出于安全的考虑，OpenEdge 提供全平台性安全证书认证模式支持，通常情况下要求所有连接 OpenEdge 的设备、应用及服务必须采用证书认证连接方式。针对 OpenEdge 各组成模块，其具体安全设计、认证策略略有不同，具体如下：

> + 针对 Hub 模块，其主要是提供设备接入能力，目前可支持 TCP、SSL（TCP + SSL）、WS(Websocket)及 WSS（Websocket + SSL）四种接入方式；
>   - 其中，针对 SSL 接入方式，OpenEdge 支持证书单向和双向认证两种模式；
> + 针对 MQTT 远程通讯模块，强烈推荐用户采用证书认证模式；
> + 针对 Agent 模块设备信息上报服务，OpenEdge 强制使用 HTTPS 安全通道协议，确保信息上报的安全；
> + 针对 Agent 模块配置下发服务，OpenEdge 同样强制使用 HTTPS 安全通道协议，确保配置下发的安全。

总体来说，在不同的模块和服务，OpenEdge 针对性地提供多种方式来确保信息交互的安全。
