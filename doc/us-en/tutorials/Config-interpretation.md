# OpenEdge Config interpretation

- [OpenEdge Config interpretation](#openedge-config-interpretation)
  - [Statement](#statement)
  - [Master Configuration](#master-configuration)
  - [Application Configuration](#application-configuration)
  - [Local Hub Configuration](#local-hub-configuration)
  - [Local Function Configuration](#local-function-configuration)
  - [MQTT Remote Configuration](#mqtt-remote-configuration)
  - [Configuration Reference](#configuration-reference)

## Statement

- The path in the configuration can be a relative path or an absolute path. If it is a relative path, it is relative to the working directory. The path configured in the container mode means the path in the container.
- Supported unit:
  - Size unit: b(byte), k(kilobyte), m(megobyte), g(gigabyte)
  - Time unit: s(second), m(minute), h(hour)

## Master Configuration

The configuration of master and application is separate. The default configuration is: [etc/openedge/openedge.yml](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/example/docker/etc/openedge/openedge.yml) of working directory. More details of master configuration as flows:

```yaml
mode: [MUST]Running mode of master module. docker: docker container mode, native: native process mode
grace: The default value is `30s`, timeout of master module gracefully exit.
api: API Server configuration of master module.
  address: default of readable environment variable: OPENEDGE_MASTER_API(address of API Server).
  timeout: The default value is `30s`, timeout of API Server.
cloud: Configuration between master module and cloud management suit, include MQTT, HTTPS.
  clientid: [MUST]mqtt client ID, must be of cloud core device.
  address: [MUST]endpoint of mqtt client connect to cloud management suit，must use ssl endpoint.
  username: [MUST]mqtt client username，must be of cloud core device.
  ca: [MUST]CA path, that the mqtt client is connected to the CA certificate of the cloud.
  key: [MUST]Client private key path, that the mqtt client connects to the client private key of the cloud.
  cert: [MUST]Client public key path, that the mqtt client connects to the client public key of the cloud.
  timeout: The default value is `30s`, means timeout of the mqtt client connects to cloud.
  interval: The default value is `1m`，means interval(doubled from 500 microseconds to maximum) of the mqtt client re-connect to cloud.
  keepalive: The default value is `1m`，means keep alive time between the mqtt client and cloud after connection has been established.
  cleansession: The default value is `false`.
  validatesubs: The default value is `false`，means whether the mqtt client checks the subscription result. If it is false, exits and return errors.
  buffersize: The default value is `10`，means the size of the memory queue sent by the mqtt client to the cloud management suit. If found exception, the mqtt client will exit and lose message.
logger: Logger configuration
  path: The default is `empty`(none configuration), that is, it does not print to the file. If the path is specified, it is output to the file(due to the path).
  level: The default value is `info`, log level, support `debug`、`info`、`warn` and `error`.
  format: The default value is `text`, log print format, support `text` and `json`.
  console: The default value is `false`, means whether print the log to terminal or not.
  age:
    max: The default value is `15`, means maximum number of days the log file is kept.
  size:
    max: The default value is `50`, log file size limit, default unit is `MB`.
  backup:
    max: The default value is `15`, the maximum number of log files to keep.
```

## Application Configuration

The default configuration of application module is: [var/db/openedge/module/module.yml](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/example/docker/var/db/openedge/module/module.yml) of working directory. More details of application configuration as flows:

```yaml
version: Application version
modules: Module list configuration
  - name: [MUST]module name, must be unique in the module list
    entry: [MUST]module entry. In the docker container mode, which means the address of image. On the contrary, which means the path of the module executable program in native process mode.
    restart: Module restart policy configuration
      retry:
        max: The default is `empty`(none configuration), which means always retry. If not, which means the maximum number of module restarts.
      policy: The default value is `always`, restart policy, support `no`, `always`, `on-failure` and `unless-stopped` four configuratons. And `no` means none restart, `always` means always restart, `on-failure` means restart the module if it exits abnormally, `unless-stopped` means the module restarts normally and restarts.
      backoff:
        min: The default value is `1s`, minimum interval of restart.
        max: The default value is `5m`, maximum interval of restart.
        factor: The default value is `2`, interval increases multiple of restart.
    expose: Port exposed configuration in Docker container mode, for example
      - 0.0.0.0:1883:1883 # If you do not want to be exposed to devices outside the host, you can set it to `127.0.0.1:1883:1883`
      - 0.0.0.0:1884:1884/tcp
      - 8080:8080/tcp
      - 8884:8884
    params: Module startup parameters, for example
      - '-c'
      - 'conf/conf.yml'
    env: Module environment variables configuration
      version: v1
    resources: [Only support docker container mode]Module resource limit configuration
      cpu:
        cpus: The percentage of CPU available to the module, for example `1.5`, means that `1.5` CPU cores can be used.
        setcpus: The CPU core available for the module, for example `0-2`, means that `0` to `2` CPU cores can be used; `0` means that the 0th CPU core can be used; `1`, which means the 1st CPU core can be used.
      memory:
        limit: The available memory of the module, for example `500m`, means that 500 megabytes of memory can be used.
        swap: The swap space available to the module, for example `1g`, means that 1G of memory can be used.
      pids:
        limit: Number of processes the module can create.
```

## Local Hub Configuration

```yaml
name: [MUST]Module name
listen: [MUST]Listening address, for example
  - tcp://0.0.0.0:1883 # In Native process mode, if you do not want to expose access to devices outside the host, you can set it to `tcp://127.0.0.1:1883`
  - ssl://0.0.0.0:1884
  - ws://:8080/mqtt
  - wss://:8884/mqtt
certificate: SSL/TLS certificate authentication configuration, if `ssl` or `wss` is enabled, it must be configured.
  ca: mqtt server CA certificate path
  key: mqtt server private key path
  cert: mqtt server public key path
principals: Access rights configuration item. If not configured, mqtt client cannot connect to the Local Hub, support username/password and certificate authentication.
  - username: username of mqtt client connect to the Local Hub
    password: password of mqtt client connect to the Local Hub
    permissions:
      - action: Operational authority. `pub` means publish permission, `sub` means subscription permission.
        permit: List of topics allowed by the operation permission, support `+` and `#` wildcards.
  - serialnumber: The serial number of the client certificate used by the mqtt client to use the certificate for mutual authentication to connect to the Local Hub.
    permissions:
      - action: Operational authority. `pub` means publish permission, `sub` means subscription permission.
        permit: List of topics allowed by the operation permission, support `+` and `#` wildcards.
subscriptions: Topic routing configuration
  - source:
      topic: subscribe topic
      qos: QoS of subscribe topic
    target:
      topic: publish topic
      qos: QoS of publish topic
message: MQTT message related configuration
  length:
    max: The default value is `32k`, which means maximum message length that can be allowed to be transmitted. The maximum can be set to 268,435,455 Byte(about 256MB).
  ingress: Message receive configuration
    qos0:
      buffer:
        size: The default value is `10000`, means the number of messages that can be cached in memory with QoS0. Increasing the cache can improve the performance of message reception. If the device loses power, it will directly discard the message with QoS0.
    qos1:
      buffer:
        size:  The default value is `100`, means the message cache size of waiting for persistent with QoS1. Increasing the cache can improve the performance of message reception, but the potential risk is that the module will exit abnormally(such as device power failure), it will lose the cached message, and will not reply(puback). The module exits normally and waits for the cached message to be processed without losing data.
      batch:
        max:  The default value is `50`, means the maximum number of messages with QoS1 can be insert into the database (persistence). After the message is persisted, it will reply with confirmation(ack).
      cleanup:
        retention:  The default value is `48h`, means the time that the message with QoS1 can be saved in the database. Messages that exceed this time will be physically deleted during cleanup.
        interval:  The default value is `1m`, means cleanup interval with QoS1.
  egress: Message publish configuration
    qos0:
      buffer:
        size:  The default value is `10000`, means the number of messages to be sent in the in-memory cache wit QoS0. If the device is powered off, the message will be discarded directly. After the buffer is full, the newly pushed message will be discarded directly.
    qos1:
      buffer:
        size:  The default value is `100`, means the size of the message buffer is not confirmed(ack) after the message with QoS1 is sent. After the buffer is full, the new message is no longer read, and the message in the cache is always acknowledged(ack). After the message with QoS1 is sent to the client, it waits for the client to confirm(puback). If the client does not reply within the specified time, the message will be resent until the client replies or the session is closed.
      batch:
        max:  The default value is `50`, means the maximum number of messages read from the database in batches.
      retry:
        interval:  The default value is `20s`, means the re-publish interval of message.
  offset: Message serial number persistence related configuration
    buffer:
      size:  The default value is `10000`, means the size of the cache queue for the serial number of the message that was acknowledged(ack). For example, three messages with QoS1 and serial numbers 1, 2, and 3 are sent to the client in batches. The client confirms the messages of sequence numbers 1 and 3. At this time, sequence number 1 will be queued and persisted. Although sequence number 3 has been confirmed, it still has to wait for the serial number 2 to be confirmed before entering the column. This design can ensure that the message can be recovered from the persistent serial number after the module restarts abnormally, ensuring that the message is not lost, but the message retransmission will occur, and therefore the message with QoS 2 is not supported.
    batch:
      max:  The default value is `100`, means the maximum number of batches of message serial numbers can be insert into the database.
logger: Logger configuration
  path: The default is `empty`(none configuration), that is, it does not print to the file. If the path is specified, it is output to the file(due to the path).
  level: The default value is `info`, log level, support `debug`、`info`、`warn` and `error`.
  format: The default value is `text`, log print format, support `text` and `json`.
  console: The default value is `false`, means whether print the log to terminal or not.
  age:
    max: The default value is `15`, means maximum number of days the log file is kept.
  size:
    max: The default value is `50`, log file size limit, default unit is `MB`.
  backup:
    max: The default value is `15`, the maximum number of log files to keep.
status: Module status configuration
  logging:
    enable: The default value is `false`, means whether to print openedge status information.
    interval: The default value is `60s`, means interval of printing openedge status information.
storage: Database storage configuration
  dir: The default value is `var/db`, means database storage directory.
shutdown: Module exit configuration
  timeout: The default value is `10m`, means timeout of module exit.
```

## Local Function Configuration

```yaml
name: [MUST]Module name
hub:
  clientid: The client ID of the mqtt client is connected to the Local Hub. If it is empty, it is randomly generated, and the clean session is forced to be true.
  address: [MUST]The address of the mqtt client is connected to the Local Hub, the address in the docker container mode is the Local Hub Module name, and the native process mode is 127.0.0.1.
  username: If you use the username/password authentication, you must fill in the username of the mqtt client that is used to connect to the Local Hub.
  password: If you use the username/password authentication, you must fill in the password of the mqtt client that is used to connect to the Local Hub.
  ca: If you use certificate mutual authentication, you must fill in the path that the mqtt client connects to the Local Hub's CA certificate.
  key: If you use certificate mutual authentication, you must fill in the path that the mqtt client connects to the Local Hub's private key.
  cert: If you use certificate mutual authentication, you must fill in the path that the mqtt client connects to the Local Hub's public key.
  timeout: The default value is `30s`, means timeout of the mqtt client connect to the Local Hub.
  interval: The default value is `1m`, means interval(doubled from 500 microseconds to maximum) of the mqtt client re-connect to Local Hub.
  keepalive: The default value is `1m`, means keep alive time between the mqtt client and the Local Hub after connection has been established.
  cleansession: The default value is `false`, means clean session status that the mqtt client connect to the Local Hub.
  validatesubs: The default value is `false`, means whether the mqtt client checks the subscription result. If it is `false`, exits and return errors.
  buffersize: The default value is `10`, means the memory queue size of that the mqtt client sends a message to the Local Hub. If the abnormal exit occurs, the message will be lost. After the recovery, the message with QoS1 depends on the Local Hub re-publish policy.
rules: Router rules configuration
  - id: [MUST]Router rule ID
    subscribe:
      topic: [MUST]means the message topic that subscribe from hub.
      qos: The default value is `0`, means the message QoS that subscribe from hub.
    compute:
      function: [MUST]means the function name that be uesd to handle message.
    publish:
      topic: [MUST]means the message topic of the hub that the function handles the output of the output message.
      qos: The default value is `0`, means the message QoS of the hub that the function handles the output of the output message.
functions:
  - name: [MUST]function name
    runtime: The runtime module name that the configuration function depends on. For `sql` runtime, that is `sql`. For `tensorflow` runtime, that is `tensorflow`. For `python2.7` runtime, that is `python27`.
    entry: The entry of the module(as above description), means the image or executable program of the runtime module running the function instance.
    handler: [MUST]handler function. For `sql` runtime, that is sql-expression, for example, `select uuid() as id, topic() as topic, * where id < 10`. For `python2.7` runtime, that is function package and handler function, for example, `sayhi.handler`. For `tensorflow` runtime, the configuration is `tag:input_tensor:output_tensor`, and the `tag` means the tag of the model, the `input_tensor` means the input node tensor name of the model network structure, the `output_tensor` the output node tensor name of the model network structure. Besides, for `tensorflow` runtime, only support saved_model network structure now, and the model must be designed to `single input, single output`.
    codedir: Optional configuration, only support `python2.7` and `tensorflow` runtime now, `sql` runtime no need. For `python2.7` runtime, means the path of the python script. For `tensorflow` runtime, means the model path, and you must set the model name to `saved_model.pb`(due to the saved_model network structure).
    env: Function environment variables configuration 
      USER_ID: acuiot
    instance: function instance configuration
      min: The default value is `0`, means the minimum number of function instance. And the minimum configuration allowed to be set is `0`, the maximum configuration allowed to be set is `100`.
      max: The default value is `1`, means the maximum number of function instance. And the minimum configuration allowed to be set is `1`, the maximum configuration allowed to be set is `100`.
      timeout: The default value is `5m`, means timeout of function instance called.
      idletime: The default value is `10m`, means the maximum idle time of a function instance is destroyed after it is exceeded. The interval for periodic check is half of idletime.
      message:
        length:
          max: The default value is `4m`, means the maximum message length allowed for function instances to be received and publish.
      cpu: [Only support docker container mode]
        cpus: The percentage of CPU available to the module, for example `1.5`, means that `1.5` CPU cores can be used.
        setcpus: The CPU core available for the module, for example `0-2`, means that `0` to `2` CPU cores can be used; `0` means that the 0th CPU core can be used; `1`, which means the 1st CPU core can be used.
      memory: [Only support docker container mode]
        limit: The available memory of the module, for example `500m`, means that 500 megabytes of memory can be used.
        swap: The swap space available to the module, for example `1g`, means that 1G of memory can be used.
      pids: [Only support docker container mode]
        limit: Number of processes the module can create.
logger: Logger configuration
  path: The default is `empty`(none configuration), that is, it does not print to the file. If the path is specified, it is output to the file(due to the path).
  level: The default value is `info`, log level, support `debug`、`info`、`warn` and `error`.
  format: The default value is `text`, log print format, support `text` and `json`.
  console: The default value is `false`, means whether print the log to terminal or not.
  age:
    max: The default value is `15`, means maximum number of days the log file is kept.
  size:
    max: The default value is `50`, log file size limit, default unit is `MB`.
  backup:
    max: The default value is `15`, the maximum number of log files to keep.
```

## MQTT Remote Configuration

```yaml
name: [MUST]Module name
hub:
  clientid: The client ID of the mqtt client is connected to the Local Hub. If it is empty, it is randomly generated, and the clean session is forced to be true.
  address: [MUST]The address of the mqtt client is connected to the Local Hub, the address in the docker container mode is the hub module name, and the native process mode is 127.0.0.1.
  username: If you use the username/password authentication, you must fill in the username of the mqtt client that is used to connect to the Local Hub.
  password: If you use the username/password authentication, you must fill in the password of the mqtt client that is used to connect to the Local Hub.
  ca: If you use certificate mutual authentication, you must fill in the path that the mqtt client connects to the Local Hub's CA certificate.
  key: If you use certificate mutual authentication, you must fill in the path that the mqtt client connects to the Local Hub's private key.
  cert: If you use certificate mutual authentication, you must fill in the path that the mqtt client connects to the Local Hub's public key.
  timeout: The default value is `30s`, means timeout of the mqtt client connect to the Local Hub.
  interval: The default value is `1m`, means interval(doubled from 500 microseconds to maximum) of the mqtt client re-connect to the Local Hub.
  keepalive: The default value is `1m`, means keep alive time between the mqtt client and the Local Hub after connection has been established.
  cleansession: The default value is `false`, means clean session status that the mqtt client connect to the Local Hub.
  validatesubs: The default value is `false`, means whether the mqtt client checks the subscription result. If it is `false`, exits and return errors.
  buffersize: The default value is `10`, means the memory queue size of that the mqtt client sends a message to the Local Hub. If the abnormal exit occurs, the message will be lost. After the recovery, the message with QoS1 depends on the Remote Hub re-publish policy.
rules: The default is empty(none configuration), means message routing rules configuration
  - id: [MUST]Router rule ID
    hub:
      subscriptions: The default is empty(none configuration), means subscribe topic(message) from Local Hub, and publish the received message to Remote Hub.
        - topic: say
          qos: 1
        - topic: hi
          qos: 0
    remote:
      name: [MUST]Remote name
      subscriptions: The default is empty(none configuration), means subscribe topic(message) from Remote Hub, and publish the received message to Local Hub.
        - topic: remote/say
          qos: 0
        - topic: remote/hi
          qos: 0
remotes: The default is empty(none configuration), means remote list
  - name: [MUST]Remote name
    clientid: The client ID of the mqtt client is connected to the Remote Hub. If it is empty, it is randomly generated, and the clean session is forced to be true.
    address: [MUST]The address of the mqtt client is connected to the Remote Hub.
    username: If you use the username/password authentication, you must fill in the username of the mqtt client that is used to connect to the Remote Hub.
    password: If you use the username/password authentication, you must fill in the password of the mqtt client that is used to connect to the Remote Hub.
    ca: If you use certificate mutual authentication, you must fill in the path that the mqtt client connects to the Remote Hub's CA certificate.
    key: If you use certificate mutual authentication, you must fill in the path that the mqtt client connects to the Remote Hub's private key.
    cert: If you use certificate mutual authentication, you must fill in the path that the mqtt client connects to the Remote Hub's public key.
    timeout: The default value is `30s`, means timeout of the mqtt client connect to the Remote Hub.
    interval: The default value is `1m`, means interval(doubled from 500 microseconds to maximum) of the mqtt client re-connect to the Remote Hub.
    keepalive: The default value is `1m`, means keep alive time between the mqtt client and the Remote Hub after connection has been established.
    cleansession: The default value is `false`, means clean session status that the mqtt client connect to the Remote Hub.
    validatesubs: The default value is `false`, means whether the mqtt client checks the subscription result. If it is `false`, exits and return errors.
    buffersize: The default value is `10`, means the memory queue size of that the mqtt client sends a message to the Remote Hub. If the abnormal exit occurs, the message will be lost. After the recovery, the message with QoS1 depends on the Local Hub re-publish policy.
logger: Logger configuration
  path: The default is `empty`(none configuration), that is, it does not print to the file. If the path is specified, it is output to the file(due to the path).
  level: The default value is `info`, log level, support `debug`、`info`、`warn` and `error`.
  format: The default value is `text`, log print format, support `text` and `json`.
  console: The default value is `false`, means whether print the log to terminal or not.
  age:
    max: The default value is `15`, means maximum number of days the log file is kept.
  size:
    max: The default value is `50`, log file size limit, default unit is `MB`.
  backup:
    max: The default value is `15`, the maximum number of log files to keep.
```

## Configuration Reference

> - [Example of Docker container mode configuration](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/example/docker/etc/openedge/openedge.yml)
> - [Example of Native process mode configuration](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/example/native/etc/openedge/openedge.yml)
