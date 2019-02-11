# OpenEdge concepts

OpenEdge is made up of **main program module, local hub module, local function module, MQTT remote module and Python2.7 runtime module.** The main capabilities of each module are as follows:

> + **Main program module** is used to manage all modules's behavior, such as start, stop, etc. And it is composed of module engine, API and cloud agent.
>   + **Module engine** controls the behavior of all modules, such as start, stop, restart, listen, etc, and currently supports **docker container mode** and **native process mode**.
>   + **Cloud agent** is responsible for the communication with **Cloud Management Suite** of [BIE](https://cloud.baidu.com/product/bie.html), and supports MQTT and HTTPS protocols. In addition, if you use MQTT protocol for communication, **must** take two-way authentication of SSL/TLS; otherwise, you **must** take one-way authentication of SSL/TLS due to HTTPS protocol.
>   + The main program exposes a set of **HTTP API**, which currently supports to start, stop and restart module, also can get free port.
> + **local hub module** is based on [MQTT](http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.html) protocol, which supports four connect modes, including **TCP**、**SSL(TCP+SSL)**、**WS(Websocket)** and **WSS(Websocket+SSL).**
> + **local function module** provides a high flexible, high available, rich scalable and quickly responsible power due to MQTT protocol. Functions are executed by one or more instances, each of them is a separate process. GRPC Server is used to run a function instance.
> + **MQTT remote module** supports MQTT protocol, can be used to synchronize messages with remote hub. In fact, it is two MQTT Server Bridge modules, which are used to subscribe to messages from one Server and forward them to the other.
> + **Python2.7 runtime module** is an implementation of **local function module**. So developers can write python script to handler messages, such as filter, exchange, forward, etc.