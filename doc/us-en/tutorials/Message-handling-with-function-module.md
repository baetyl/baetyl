# Message handling with Local Function Service

**Statement**

- The operating system as mentioned in this document is Darwin.
- The version of runtime is Python3.6, and for Python2.7, configuration is the same except fot the language difference when coding the scripts
- The MQTT client toolkit as mentioned in this document is [MQTTBOX](../Resources-download.md#mqttbox-download).
- The docker image used in this document is compiled from the OpenEdge source code. More detailed contents please refer to [Build OpenEdge from source](../setup/Build-OpenEdge-from-Source.md).
- In this article, the service created based on the Hub module is called `localhub` service.

Different from the `localhub` service to transfer message among devices(mqtt clients), this document describes the message handling with Local Function Manager service(also include Local Hub service and Python3.6 runtime service). In the document, Local Hub service is used to establish connection between OpenEdge and mqtt client, Python3.6 runtime service is used to handle MQTT messages, and the Local Function Manager service is used to combine `localhub` service with Python3.6 runtime service with message context.

This document will take the TCP connection method as an example to show the message handling, calculation and forwarding with Local Function Manager service.

## Workflow

- Step 1：Startup OpenEdge in docker container mode.
- Step 2：MQTTBOX connect to `localhub` Service by TCP connection method, more detailed contents please refer to [Device connect to OpenEdge with Hub Module](./Device-connect-to-OpenEdge-with-hub-module.md)
    - If connect successfully, then subscribe the MQTT topic due to the configuration of `localhub` Service, and observe the log of OpenEdge.
        - If the OpenEdge's log shows that the Python Runtime Service has been started, it indicates that the published message was handled by the specified function.
        - If the OpenEdge's log shows that the Python Runtime Service has not been started, then retry it until the Python Runtime Service has been started.
    - If connect unsuccessfully, then retry `Step 2` operation until it connect successfully
- Step 3：Check the publishing and receiving messages via MQTTBOX.

![Workflow of using Local Function Manager Service to handle MQTT messages](../../images/tutorials/process/openedge-python-flow.png)

## Message Handling Test

The configuration of the `localhub` Service and the Local Function Manager Service used in the test is as follows:

```yaml
# The configuration of localhub service
listen:
  - tcp://0.0.0.0:1883
principals:
  - username: 'test'
    password: 'hahaha'
    permissions:
      - action: 'pub'
        permit: ['#']
      - action: 'sub'
        permit: ['#']

# The configuration of Local Function Manager service
hub:
  address: tcp://localhub:1883
  username: test
  password: hahaha
rules:
  - clientid: localfunc-1
    subscribe:
      topic: t
    function:
      name: sayhi
    publish:
      topic: t/hi
functions:
  - name: sayhi
    service: function-sayhi
    instance:
      min: 0
      max: 10
      idletime: 1m

# The configuration of Python3.6 runtime
functions:
  - name: 'sayhi'
    handler: 'sayhi.handler'
    codedir: 'var/db/openedge/function-sayhi'

# The configuration of application.yml
version: v0
services:
  - name: localhub
    image: openedge-hub
    replica: 1
    ports:
      - 1883:1883
    mounts:
      - name: localhub-conf
        path: etc/openedge
        readonly: true
      - name: localhub-data
        path: var/db/openedge/data
      - name: localhub-log
        path: var/log/openedge
  - name: function-manager
    image: openedge-function-manager
    replica: 1
    mounts:
      - name: function-manager-conf
        path: etc/openedge
        readonly: true
      - name: function-manager-log
        path: var/log/openedge
  - name: function-sayhi
    image: openedge-function-python36
    replica: 0
    mounts:
      - name: function-sayhi-conf
        path: etc/openedge
        readonly: true
      - name: function-sayhi-code
        path: var/db/openedge/function-sayhi
        readonly: true
volumes:
  # hub
  - name: localhub-conf
    path: var/db/openedge/localhub-conf
  - name: localhub-data
    path: var/db/openedge/localhub-data
  - name: localhub-log
    path: var/db/openedge/localhub-log
  # function manager
  - name: function-manager-conf
    path: var/db/openedge/function-manager-conf
  - name: function-manager-log
    path: var/db/openedge/function-manager-log
  # function python runtime sayhi
  - name: function-sayhi-conf
    path: var/db/openedge/function-sayhi-conf
  - name: function-sayhi-code
    path: var/db/openedge/function-sayhi-code
```

The directory of configuration tree is as follows:

```shell
var/
└── db
    └── openedge
        ├── application.yml
        ├── function-manager-conf
        │   └── service.yml
        ├── function-sayhi-code
        │   ├── __init__.py
        │   └── sayhi.py
        ├── function-sayhi-conf
        │   └── service.yml
        └── localhub-conf
            └── service.yml
```

As configured above, if the MQTTBOX has established a connection with OpenEdge via the `localhub` Service, the message published to the topic `t` will be handled by `sayhi` function, and the result will be published to the `localhub` Service with the topic `t/hi`. At the same time, the MQTT client subscribed the topic `t/hi` will receive the result message.

_**NOTE**: Any function that appears in the `rules` configuration must be configured in the `functions` configuration, otherwise OpenEdge will not be started._

### OpenEdge Start

As described in `Step 1`, OpenEdge starts with docker container mode. And it can be found that OpenEdge is starting via OpenEdge's log. More detailed contents are as follows:

![OpenEdge start](../../images/tutorials/process/openedge-function-start.png)

Also, we can execute the command `docker ps` to view the list of docker containers currently running.

![View the list of docker containers currently running](../../images/tutorials/process/openedge-docker-ps-after.png)

After comparison, it is not difficult to find that the two container modules of the `localhub` Service and the Local Function Manager Service have been successfully loaded at the time of OpenEdge startup.

### MQTTBOX Establish a Connection with OpenEdge

Here, using MQTTBOX as the MQTT client, click the `Add subscriber` button to subscribe the topic `t/hi`. And topic `t/hi` is used to receive the result message after `sayhi` function handled. More detailed contents are as shown below.

![MQTTBOX connection configuration](../../images/tutorials/process/mqttbox-tcp-process-config.png)

The figure above shows that MQTTBOX has successfully subscribed the topic `t/hi`.

### Message Handling Check

Based on the above, here we use the Python function `sayhi` to handle the message of the topic `t` and publish the result back to the topic `t/hi`. More detailed contents of function `sayhi` are as shown below.

```python
#!/usr/bin/env python36
#-*- coding:utf-8 -*-
"""
service to say hi
"""

import os

def handler(event, context):
    """
    function handler
    """
    if 'USER_ID' in os.environ:
      event['USER_ID'] = os.environ['USER_ID']

    if 'functionName' in context:
      event['functionName'] = context['functionName']

    if 'functionInvokeID' in context:
      event['functionInvokeID'] = context['functionInvokeID']

    if 'invokeid' in context:
      event['invokeid'] = context['invokeid']

    if 'messageQOS' in context:
      event['messageQOS'] = context['messageQOS']

    if 'messageTopic' in context:
      event['messageTopic'] = context['messageTopic']

    event['py'] = '你好，世界！'

    return event
```

It can be found that after receiving a message in a dictionary(`dict`) format, the function `sayhi` will handle it and then return the result. The returned result include: environment variable `USER_ID`, function name `functionName`, function call ID `functionInvokeID`, input message subject `messageTopic`, input message message QoS `messageQOS` and other fields.

Here, we publish the message `{"id":10}` to the topic `t` via MQTTBOX, and then observe the receiving message of the topic `t/hi`. More detailed contents are as shown below.

![MQTTBOX successfully receive the result via topic `t/hi`](../../images/tutorials/process/mqttbox-tcp-process-success.png)

It is not difficult to find that the result received by MQTTBOX via the topic `t/hi` is consistent with the above analysis.

In addition, we can observe the OpenEdge's log and execute the command `docker ps` again to view the list of containers currently running on the system. The results are shown below.

![OpenEdge's log when using python runtime](../../images/tutorials/process/openedge-python-start.png)

![View the list of docker containers currently running](../../images/tutorials/process/openedge-docker-ps-python-start.png)

As you can see from the above two figures, except the `localhub` Service and the Local Function Manager Service were loaded when OpenEdge started, the Python Runtime Service was also loaded when the MQTT message of topic `t` was handled by function `sayhi`. More detailed designed contents of Python Runtime Service please refer to [OpenEdge design](../overview/OpenEdge-design.md).