# Message Synchronize between OpenEdge and Baidu IoT Hub via Remote Module

**Statement**

> + The operating system as mentioned in this document is Darwin.
> + The MQTT client toolkit as mentioned in this document are [MQTTBOX](../Resources-download.md#mqttbox-download) and [MQTT.fx](../Resources-download.md#mqtt.fx-download).
> + The docker image used in this document is compiled from the OpenEdge source code. More detailed contents please refer to [Build OpenEdge from source](../setup/Build-OpenEdge-from-Source.md)
> + The Remote Hub as mentioned in this document is [Baidu IoT Hub](https://cloud.baidu.com/product/iot.html)

The Remote Module was developed to meet the needs of the IoT scenario. The OpenEdge(via Local Hub Module) can synchronize message with Remote Hub services(such as[Azure IoT Hub](https://azure.microsoft.com/en-us/services) /iot-hub/), [AWS IoT Core](https://amazonaws-china.com/iot-core/), [Baidu IoT Hub](https://cloud.baidu.com/product/iot.html), etc.) via the Remote Module. That is to say, through the Remote Module, we can either subscribe the message from Remote Hub and publish it to the Local Hub Module or subscribe the message from Local Hub Module and publish it to Remote Hub service. The configuration of Remote Module can refer to [Remote Module Configuration](./Config-interpretation.md#mqtt-remote-configuration).

## Workflow

- `Step 1`：Create device(MQTT client) connection info(include `endpoint`, `user`, `principal`, `policy`, etc.) via Baidu IoT Hub.
- `Step 2`：Select MQTT.fx as the MQTT client that used to connect to Baidu IoT Hub.
  - If connect successfully, then do the following next.
  - If connect unsuccessfully, then retry it until it connect successfully. More detailed contents can refer to [How to connect to Baidu IoT Hub via MQTT.fx](https://cloud.baidu.com/doc/IOT/GettingStarted.html#.E6.95.B0.E6.8D.AE.E5.9E.8B.E9.A1.B9.E7.9B.AE)。
- `Step 3`：Startup OpenEdge in docker container mode, and observe the log of OpenEdge.
  - If the Local Hub Module and Remote Module start successfully, then do the following next.
  - If the Local Hub Module and Remote Module start unsuccessfully, then retry `Step 3` until they start successfully.
- `Step 4`：Select MQTTBOX as the MQTT client that used to connect to the Local Hub.
    - If connect successfully, then do the following next.
    - If connect unsuccessfully, then retry `Step 4` until it connect successfully.
- `Step 5`：Due to the configuration of Remote Module, using MQTTBOX publish message to the specified topic, and observing the receiving message via MQTT.fx. Similarly, using MQTT.fx publish message to the specified topic, and observing the receiving message via MQTTBOX.
- `Step 6`：If both parties in `Step 5` can receive the message content posted by the other one, it indicates the Remote function test passes smoothly.

The workflow diagram is as follows.

![using Remote Module to synchronize message](../../images/tutorials/remote/openedge-remote-flow.png)

## Message Synchronize via Remote Module

Firstly, the configuration of the Remote Module used in the document is as follows.

```yaml
name: openedge-remote-mqtt
hub:
  address: tcp://openedge-hub:1883
  username: test
  password: hahaha
remotes:
  - name: remote
    address: tcp://u4u6zk2.mqtt.iot.bj.baidubce.com:1883
    clientid: 349360d3c91a4c55a57139e9085e526f
    username: u4u6zk2/demo
    password: XqySIYMBsjK0JkEh
rules:
  - id: rule-rcg3k6ytq
    hub:
      subscriptions:
        - topic: t1
          qos: 1
    remote:
      name: remote
      subscriptions:
        - topic: t2
          qos: 1
```

According to the configuration of the above, it means that the Remote Module subscribes the topic `t1` from the Local Hub Module, subscribes the topic `t2` from Baidu IoT Hub. When MQTTBOX publishes a message to the topic `t1`, the Local Hub Module will receive this message and forward it to Baidu IoT Hub via Remote Module, and MQTT.fx will also receive this message(suppose MQTT.fx has already subscribed the topic `t1` before) from Baidu IoT Hub. Similarly, When we use MQTT.fx to publish a message to the topic `t2`, then Baidu IoT Hub will receive it and forward it to the Local Hub Module via Remote module. Finally, MQTTBOX will receive this message(suppose MQTTBOX has already subscribed the topic `t2` before).

In a word, from MQTTBOX publishes a message to the topic `t1`, to MQTT.fx receives the message, the routing path of the message is as follows.

> **MQTTBOX -> Local Hub Module -> Remote Module -> Baidu IoT Hub -> MQTT.fx**

Similarly, from MQTT.fx publishes a message to the topic `t2`, to MQTTBOX receives the message, the routing path of the message is as follows.

> **MQTT.fx -> Baidu IoT Hub -> Remote Module -> Local Hub Module -> MQTTBOX**

### Establish a Connection between MQTT.fx and Baidu IoT Hub

As described in `Step 1, Step 2`, the detailed contents of the connection between MQTT.fx and Baidu IoT Hub are as follows.

![Create `endpoint` via Baidu IoT Hub](../../images/tutorials/remote/cloud-iothub-config.png)

![Create other information via Baidu IoT Hub](../../images/tutorials/remote/cloud-iothub-user-config.png)

![Configuration of MQTT.fx](../../images/tutorials/remote/mqttfx-connect-hub-config.png)

After set the configuration of MQTT.fx, click `OK` or `Apply` button, then click `Connect` button, and wait for the connecting. Also, we can check if the connection status is OK via the color button. When the button's color change to **Green**, that is to say, the connection is established. More detailed contents are shown below.

![Successfully establish a connection between MQTT.fx and Baidu IoT Hub](../../images/tutorials/remote/mqttfx-connect-success.png)

After the connection is established, switch to the `Subscribe` page and subscribe the topic `t1`. More detailed contents are shown below.

![MQTT.fx successfully subscribe the topic `t1`](../../images/tutorials/remote/mqttfx-sub-t1-success.png)

### Establish a Connection between MQTTBOX and the Local Hub Module

As described in `Step 3`, the Local Hub Module and Remote Module also loaded when OpenEdge started. More detailed contents are shown below.

![OpenEdge successfully load Hub、Remote](../../images/tutorials/remote/openedge-hub-remote-start.png)

In addition, we can execute the command `docker ps` to view the list of docker containers currently running on the system.

![View the list of docker containers currently running](../../images/tutorials/remote/openedge-docker-ps-hub-remote-run.png)

After OpenEdge successfully startup, set the configuration of connection, then establish the connection with the Local Hub Module and subscribe the topic `t2`.

![MQTTBOX successfully subscribe the topic `t2`](../../images/tutorials/remote/mqttbox-sub-t2-success.png)

### Message Synchronize Test

Here, MQTT.fx and MQTTBOX will be used as message publishers, and the other one will be used as a message receiver.

**MQTT.fx publishes message, and MQTTBOX receives message**

Firstly, using MQTT.fx publishes a message `This message is from MQTT.fx.` to the topic `t2`.

![Pbulishing a message to the topic `t2` via MQTT.fx](../../images/tutorials/remote/mqttfx-pub-t2-success.png)

At the same time, observing the message receiving status of MQTTBOX via the topic `t2`.

![MQTTBOX successfully received the message](../../images/tutorials/remote/mqttbox-receive-t2-message-success.png)

**MQTTBOX publishes message, and MQTT.fx receives message**

Similarly, publishing the message `This message is from MQTTBOX.` to the topic `t1` via MQTTBOX.

![Publishing a message to the topic `t1` via MQTTBOX](../../images/tutorials/remote/mqttbox-pub-t1-success.png)

Then we can observe the message receiving status of MQTT.fx via the topic `t1`.

![MQTT.fx successfully received the message](../../images/tutorials/remote/mqttfx-receive-t1-message-success.png)

In summary, both MQTT.fx and MQTTBOX have correctly received the specified message, and the content is consistent.