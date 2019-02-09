# OpenEdge

OpenEdge contains a master program and some modules. The master program manages all modules through configuration. Currently, OpenEdge supports two modes of running, namely **docker** container mode and **native** process mode.

Docker Containr Mode Design Diagram:

![Docker Containr Design Diagram](../../images/overview/design/mode_docker.png)

Native Process Design Diagram:

![Native Process Design Diagram](../../images/overview/design/mode_native.png)

## Master

[Master](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/master/master.go) is responsible for the management of all modules, cloud synchronization, etc., and is composed of a module engine, a cloud agent, and an API server.

### Engine

[Engine](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/engine/engine.go)
is responsible for the management of all modules includes operations of start, stop and supervise. Currently supports docker container mode and native process mode.

Engine loads the list of modules from the [var/db/openedge/module/module.yml](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/example/docker/var/db/openedge/module/module.yml) configuration in the working directory and starts the modules one by one in the order of the list. Engine will start a daemon coroutine for each module to monitor the module status. If the module exits abnormally, it will restart or exit according to the module's [Restart Policy](../tutorials/local/Config-interpretation.md#应用配置). When master is closing, engine will close all modules one by one in the reverse order of the list.

_**TIP**: The working directory can be specified by -w when OpenEdge is started. The default is the directory above the directory where OpenEdge executables are located._

In docker container mode, the module is started as a docker container by the docker client using the docker image specified by the entry configuration, and automatically joins the custom docker network (openedge). Since the docker comes with a DNS server, the modules can communicate with each other through the module name. In addition, the module can expose the port through the expose configuration; the resources that can be used by the module are restricted by the resources configuration, and currently the CPU, memory, and process limit are supported. [Configuration Reference](../tutorials/local/Config-interpretation.md#应用配置)

In docker container mode, the working directory of the module in the container is the root directory: /; the configuration path is: /etc/openedge/module.yml; the resource files directory is: /var/db/openedge/module/<module name>; the persistent data output directory is: /var/db/openedge/volume/<module name>; the log output directory is: /var/log/openedge/<module name>. The docker volumes mapping is as follows:

> - \<openedge_host_work_dir>/var/db/openedge/module/<module_name>/module.yml:/etc/openedge/module.yml
> - \<openedge_host_work_dir>/var/db/openedge/module/<module_name>:/var/db/openedge/module/<module_name>
> - \<openedge_host_work_dir>/var/db/openedge/volume/<module_name>:/var/db/openedge/volume/<module_name>
> - \<openedge_host_work_dir>/var/log/openedge/<module_name>:/var/log/openedge/<module_name>

_**TIP**: The mark value will change after the module configuration is modified in the cloud, similar to the revision of git._

In native process mode, the module is started by syscall using the executable file specified by the entry configuration. The modules can communicate with each other through localhost (127.0.0.1). **Does not support resource isolation and resource limitation in this mode**.

### Agent

The cloud agent is responsible for communicating with the cloud management suite through the MQTT and HTTPS channels, MQTT forcing the two-way authentication of the SSL/TLS certificate, and HTTPS forcing the one-way authentication of the SSL/TLS certificate. [Cloud Configuration Reference](../tutorials/local/Config-interpretation.md#主程序配置)。

After OpenEdge startup and hot loading are completed, the device information is reported through the cloud agent. The current [reported content](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/agent/report.go) is as follows:

> - go_version: Golang version of OpenEdge master program
> - bin_version: the release version of the OpenEdge master program
> - conf_version: the configuration version loaded by the OpenEdge master program
> - error: exception information, if any
> - os: device system, such as: linux, windows, darwin
> - bit: device CPU digits, for example: 32, 64
> - arch: device CPU architecture, for example: amd64
> - gpu\<n\>: the model information of the nth GPU of the device
> - gpu\<n\>_mem_total: total memory capacity of the nth GPU of the device
> - gpu\<n\>_mem_free: the remaining memory capacity of the nth GPU of the device
> - mem_total: total memory capacity of the device
> - mem_free: the remaining memory capacity of the device
> - swap_total: total swap space of the device
> - swap_free: the remaining capacity of the swap space of the device

After the cloud agent receives the reload event from the cloud management suite, OpenEdge starts executing [hot loading](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/master/master.go)，The process is as follows:

![Hot loading process](../../images/overview/design/reload.png)

_**NOTE**: At present, the hot load adopts the method of full update, which stops all the old application modules and then starts all new application modules, so the service provided by the module will be interrupted. In addition, decompressing the new application package will directly overwrite the old application files, and the redundant old application files will not be cleaned up._

_**TIP**: If the device can't connect to the external network or leave the cloud management, you can remove the cloud configuration item and run it offline._

### API

The OpenEdge master program exposes a set of HTTP APIs that currently support getting idle ports, starting and stopping modules. API server listens to unix domain socket by default on linux system, var/openedge.sock under working directory; listens to TCP address on other systems, and tcp://127.0.0.1:50050 is the default address. Developers can modify the listening address via the [api](../tutorials/local/Config-interpretation.md#主程序配置) configuration. In order to facilitate management, we have made a division of the module. The module loaded from var/db/openedge/module/module.yml is called the resident module. The module started by resident module through the API is called the temporary module which should be stopped by the same resident module. When OpenEdge exits, all resident modules will be stopped in reverse order. The resident module will also call the API to stop the module it starts. In the end, if there are missing temporary modules, they will be stopped in randomly order. To access the API, you need to provide an account username and password. Set the following two request headers:

> - x-iot-edge-username: account username which is the resident module name
> - x-iot-edge-password: account password which is the resident module token, obtained from the environment variable using ```module.GetEnv(module.EnvOpenEdgeModuleToken)```. A temporary token is generated for each resident module when the master program starts.

The official function module will call this API to start and stop the function instance (runtime) module. When the function module stops, it is responsible for stopping all the temporary modules that it starts.

### ENV

OpenEdge sets the following system environment variables for all modules currently:

> - OPENEDGE_HOST_OS: System type of the device (host) where OpenEdge is located
> - OPENEDGE_MASTER_API: API address of the OpenEdge master program
> - OPENEDGE_MODULE_MODE: Running mode of OpenEdge master program
> - OPENEDGE_MODULE_TOKEN: The temporary token assigned to the resident module by the OpenEdge master program, which can be used as account password to access the API of the master program.

The official function module is to connect to the OpenEdge master program by reading OPENEDGE_MASTER_API. For example, the OPENEDGE_MASTER_API on Linux system is unix://var/openedge.sock by default; on other systems in docker container mode is tcp://host.docker.internal:50050 by default; on other systems in native process mode is  tcp://127.0.0.1:50050 by default.

_**NOTE**: Environment variables configured in the application will be overwritten if they are the same as the above system environment variables._

## Official Modules

At present, the official provides some modules to meet some common application scenarios, of course, developers can also develop their own modules, as long as they meet the loading requirements of custom modules.

### openedge-hub

The Hub module is a stand-alone version of the message subscription and distribution center, using the MQTT 3.1.1 protocol. The design diagram is as follows:

![Hub Module Design Diagram](../../images/overview/design/hub.png)

Currently supports 4 access methods: tcp, ssl (tcp+ssl), ws (websocket) and wss (websocket+ssl). The MQTT protocol support is as follows:

> - Supports connect, disconnect, subscribe, publish, unsubscribe, ping, etc.
> - Support for message publishing and subscriptions with QoS levels 0 and 1
> - Support retain, will message, clean session
> - Support for subscribing to topics with wildcards such as "+" and "#"
> - Supports validation of conforming client id and payload
> - **Not support** to publish and subscribe to topics prefixed with "$"
> - **Not support** client's keep alive feature and QoS level 2

_**NOTE**:_

> - The maximum number of separators "/" in the publish and subscribe topic is no more than 8, and the maximum length of the topic name is no more than 255 characters.
> - The default maximum length of message packets is 32k. The maximum length that can be supported is 268,435,455 (byte), less than 256m, which can be modified by the message configuration.
> - clientid supports uppercase and lowercase letters, numbers, underscores, hyphens (minus sign), and null characters (the null character indicates that the client is a temporary connection, forcing cleansession=true), and the maximum length is no more than 128 characters.
> - The QoS of the message can only be dropped. For example, when the QoS of the original message is 0, even if the subscription QoS is 1, the message is sent at the level of QoS 0.

The Hub module supports simple topic routing, such as subscribing to a message with the topic 't' and publish it back to the broker with a new topic 't/topic'. [Configuration Reference](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/example/docker/var/db/openedge/module/localhub/module.yml)

#### openedge-function

Function module provides the computing power based on MQTT message mechanism, flexibility, high availability, good scalability and fast response, and is compatible with [Baidu CFC](https://cloud.baidu.com/product/cfc.html). The function is executed by one or more concrete instances, each of which is a separate process. The function instance is now running as the grpc server. All function instances are managed by the instance pool and support automatic scale-up and scale-down. The design diagram is as follows:

![Function Module Design Diagram](../../images/overview/design/function.png)

_**NOTE**: If the function is executed incorrectly, the function module returns a message of the following format for subsequent processing. Where packet is the message input by the function (the message being processed), not the message returned by the function._

```python
{
    "errorMessage": "rpc error: code = Unknown desc = Exception calling application",
    "errorType": "*errors.Err",
    "packet": {
        "Message": {
            "Topic": "t",
            "Payload": "eyJpZCI6MSwiZGV2aWNlIjoiMTExIn0=",
            "QOS": 0,
            "Retain": false
        },
        "Dup": false,
        "ID": 0
    }
}
```

#### openedge-function-runtime-python27

The Python function is similar to [Baidu CFC](https://cloud.baidu.com/product/cfc.html). Users can process messages by writing their own functions, which can filter, convert, forward messages, etc., the use is very flexible.

The input and output of a Python function can be either JSON or binary. The message payload will try JSON decoding (json.loads(payload)) before passing it as a parameter. If it succeeds, it will pass the dictionary (dict) type. If it fails, it will pass the original binary data.

Python function supports reading environment variables such as os.environ['PATH'].

Python function supports reading contexts such as context['functionName'].

Python function implementation example:

```python
#!/usr/bin/env python
#-*- coding:utf-8 -*-
"""
module to say hi
"""

def handler(event, context):
    """
    function handler
    """
    event['functionName'] = context['functionName']
    event['functionInvokeID'] = context['functionInvokeID']
    event['functionInstanceID'] = context['functionInstanceID']
    event['messageQOS'] = context['messageQOS']
    event['messageTopic'] = context['messageTopic']
    event['sayhi'] = '你好，世界！'
    return event
```

_**TIP**: In the native process mode, to run [sayhi.py](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/example/native/var/db/openedge/module/func-nyeosbbch/sayhi.py) provided by this codebase, you need to install python2.7 yourself, and you need to install protobuf3, grpcio, pyyaml based on python2.7 (`pip install grpcio protobuf pyyaml`)._

In addition, for the construction of the native process mode python script runtime environment, it is recommended to build a virtual environment through virtualenv, and install related dependencies on this basis, the relevant steps are as follows:

> ```python
> # install virtualenv via pip
> [sudo] pip install virtualenv
> # test your installation
> virtualenv --version
> # build workdir run environment
> cd /path/to/native/workdir
> virtualenv native/workdir
> # install requirements
> source bin/activate # activate virtualenv
> pip install grpcio protobuf pyyaml # install grpc, protobuf3, yaml via pip
> # test native mode
> bin/openedge # run openedge under native mode
> deactivate # deactivate virtualenv
> ```

### openedge-remote-mqtt

The remote communication module currently supports the MQTT protocol, which can bridge two MQTT servers for subscribing messages from one server message and publishing it to another server. Currently, multiple remote and hub can be configured to synchronize messages at the same time. The design diagram is as follows:

![Remote Module Design Diagram](../../images/overview/design/remote.png)

As shown in the figure above, here, the OpenEdge remote communication module (openedge-remote-mqtt) is used to forward and synchronize messages between the OpenEdge local Hub module and the remote cloud Hub platform.
