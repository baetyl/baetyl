# OpenEdge Concepts

OpenEdge is made up of **Master Program, Local Hub Module, Local Function Module, MQTT Remote Module and Python Runtime Module.** The main capabilities of each module are as follows:

> + **Master Program** is used to manage all modules's behavior, such as start, stop, etc. And it is composed of module engine, API and cloud agent.
>   + **Module Engine** controls the behavior of all modules, such as start, stop, restart, listen, etc, and currently supports **Docker Container Mode** and **Native Process Mode**.
>   + **Cloud Agent** is responsible for the communication with **Cloud Management Suite** of [BIE](https://cloud.baidu.com/product/bie.html), and supports MQTT and HTTPS protocols. In addition, if you use MQTT protocol for communication, **MUST** take two-way authentication of SSL/TLS; otherwise, you **MUST** take one-way authentication of SSL/TLS due to HTTPS protocol.
>   + The main program exposes a set of **HTTP API**, which currently supports to start, stop and restart module, also can get free port.
> + **Local Hub Module** is based on [MQTT](http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.html) protocol, which supports four connect modes, including **TCP**、**SSL(TCP+SSL)**、**WS(Websocket)** and **WSS(Websocket+SSL)**.
> + **Local Function Module** provides a high flexible, high available, rich scalable and quickly responsible power due to MQTT protocol. Functions are executed by one or more instances, each of them is a separate process. GRPC Server is used to run a function instance.
> + **MQTT Remote Module** supports MQTT protocol, can be used to synchronize messages with remote hub. In fact, it is two MQTT Server Bridge modules, which are used to subscribe to messages from one Server and forward them to the other.
> + **Python2.7 Runtime Module** is an implementation of **Local Function Module**. So developers can write python script to handler messages, such as filter, exchange, forward, etc.