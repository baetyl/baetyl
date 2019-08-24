# Device connect to Baetyl with Hub module

**Statement**:

- The device system used in this test is Ubuntu18.04
- It should be installed for Baetyl when you read this document, more details please refer to [How-to-quick-install-Baetyl](../setup/Quick-Install.md)
- MQTT.fx and MQTTBOX are MQTT Clients in this test, which [MQTT.fx](../Resources-download.md) used for TCP and SSL connection test and [MQTTBOX](../Resources-download.md) used for WS (Websocket) connection test.
- The hub module image used  is the official image published in the Baetyl Cloud Management Suite: `hub.baidubce.com/baetyl/baetyl-hub:latest`
- You can also compile the required Hub module image by using Baetyl source code. Please see [How to build image from source code](../setup/Build-Baetyl-from-Source.md)

The complete configuration reference for [Hub Module Configuration](./Config-interpretation.md).

**NOTE**：Darwin can install Baetyl by using Baetyl source code. Please see [How to build image from source code](../setup/Build-Baetyl-from-Source.md).

## Workflow

- Step 1: Write the configuration according to the usage requirements, and then execute `sudo systemctl start baetyl`(you should install Baetyl first, more detailed contents can refer to [How-to-quick-install-Baetyl](../setup/Quick-Install.md)) to start the Baetyl in Docker container mode. Then execute the command `sudo systemctl status baetyl` to check whether baetyl is running.
- Step 2: Configure the MQTT Client according to the connection protocol selected.
    - If TCP protocol was selected, you only need to configure the username and password(see the configuration option username and password of principals) and fill in the corresponding port.
    - If SSL protocol was selected, username, private key, certificate and CA should be need. then fill in the corresponding port;
    - If WS protocol was selected, you only need to configure the username, password, and corresponding port.
- Step 3: Step 3: If all the above steps are normal and operations are correct, you can check the connection status through the log of Baetyl or MQTT Client.

**NOTE**：In the latest version of Baetyl Hub Module, password has been changed to plaintext storage.

## Connection Test

As mentioned above, Baetyl must be started before the connection test.

### Baetyl Connection Test

Configuration file location for the Baetyl main program is: `var/db/baetyl/application.yml`.

The configuration of Baetyl Master are as follows:

```yaml
version: v0
services:
  - name: localhub
    image: hub.baidubce.com/baetyl/baetyl-hub:latest
    replica: 1
    ports:
      - 1883:1883
      - 8883:8883
      - 8080:8080
    mounts:
      - name: localhub-conf
        path: etc/baetyl
        readonly: true
      - name: localhub-cert
        path: var/db/baetyl/cert
        readonly: true
      - name: localhub-data
        path: var/db/baetyl/data
      - name: localhub-log
        path: var/log/baetyl
volumes:
  - name: localhub-conf
    path: var/db/baetyl/localhub-conf
  - name: localhub-data
    path: var/db/baetyl/localhub-data
  - name: localhub-cert
    path: var/db/baetyl/localhub-cert-only-for-test
  - name: localhub-log
    path: var/db/baetyl/localhub-log
```

Configuration file location for the Baetyl Hub module is: `var/db/baetyl/localhub-conf/service.yml`.

The configuration of Baetyl Hub Module are as follows:

```yaml
listen:
  - tcp://0.0.0.0:1883
  - ssl://0.0.0.0:8883
  - ws://0.0.0.0:8080/mqtt
certificate:
  ca: var/db/baetyl/cert/ca.pem
  cert: var/db/baetyl/cert/server.pem
  key: var/db/baetyl/cert/server.key
principals:
  - username: two-way-tls
    permissions:
      - action: 'pub'
        permit: ['tls/#']
      - action: 'sub'
        permit: ['tls/#']
  - username: test
    password: hahaha
    permissions:
      - action: 'pub'
        permit: ['#']
      - action: 'sub'
        permit: ['#']
subscriptions:
  - source:
      topic: 't'
    target:
      topic: 't/topic'
logger:
  path: var/log/baetyl/service.log
  level: "debug"
```

### Baetyl Startup

According to Step 1, execute `sudo systemctl start baetyl` to start Baetyl in Docker mode and then execute the command `sudo systemctl status baetyl` to check whether baetyl is running. The normal situation is shown as below.

![Baetyl status](../../images/setup/baetyl-systemctl-status.png)

**NOTE**：Darwin can install Baetyl by using Baetyl source code, and excute `sudo baetyl start` to start the Baetyl in Docker container mode.

![Baetyl startup](../../images/tutorials/connect/baetyl-hub-start.png)

As you can see, the image of Hub module has been loaded after Baetyl starts up normally. Alternatively, you can use `docker ps` command to check which docker container is currently running.

![docker ps](../../images/tutorials/connect/container-baetyl-hub-run.png)

Container mode requires port mapping, allowing external access to the container, the configuration item is the `ports` field in the main program configuration file.

As mentioned above, when the Hub Module starts, it will open ports 1883, 8883 and 8080 at the same time, which are used for TCP, SSL, WS (Websocket) protocol. Then we will use MQTTBOX and MQTT.fx as MQTT client to check the connection between MQTT client and Baetyl.

**TCP Connection Test**

Startup MQTT.fx, enter the `Edit Connection Profiles` page, fill in the `Profile Name`, `Broker Address` and `Port` according to the connection configuration of Baetyl Hub module, and then configure the `username` & `password` in User Credentials according to the `principals` configuration. Then click `Apply` button to complete the connection configuration of MQTT.fx with TCP protocol.

![TCP connection configuration](../../images/tutorials/connect/mqttbox-tcp-connect-config.png)

Then close the configuration page, select the Profile Name configured, then click `Connect` button, if the connection configuration information matches the `principals` configuration of Baetyl Hub module, you can see the connection success flag which as shown below.

![TCP connection success](../../images/tutorials/connect/mqttbox-tcp-connect-success.png)

**SSL Connection Test**

Startup MQTT.fx and enter the Edit Connection Profiles page. Similar to the TCP connection configuration, fill in the profile name, broker address, and port. For SSL protocol, you need to fill in the username in `User Credentials` and configure SSL/TLS option as shown below. Then click the `Apply` button to complete the connection configuration of MQTT.fx in SSL connection method.

![SSL connection configuration1](../../images/tutorials/connect/mqttbox-ssl-connect-config1.png)

![SSL connection configuration2](../../images/tutorials/connect/mqttbox-ssl-connect-config2.png)

Then close the configuration page, select the Profile Name configured, then click `Connect` button, if the connection configuration information matches the `principals` configuration of Baetyl Hub module, you can see the connection success flag which as shown below.

![SSL connection success](../../images/tutorials/connect/mqttbox-ssl-connect-success.png)

**WS (Websocket) Connection Test**

Startup MQTTBOX, enter the Client creation page, select the `ws` protocol, configure the broker address and port according to the Baetyl Hub module, fill in the username and password according to the `principals` configuration option, and click the `save` button. Then complete the connection configuration of MQTTBOX in WS connection method which as shown below.

![WS（Websocket）connection configuration](../../images/tutorials/connect/mqttbox-ws-connect-config.png)

Once the above operation is correct, you can see the sign of successful connection with Baetyl Hub in MQTTBOX, which is shown in the figure as below.

![WS（Websocket）connection success](../../images/tutorials/connect/mqttbox-ws-connect-success.png)

In summary, we successfully completed the connection test for the Baetyl Hub module through MQTT.fx and MQTTBOX. In addition, we can also write test scripts to connect to Baetyl Hub through Paho MQTT. For details, please refer to [Related Resources Download](../Resources-download.md).
