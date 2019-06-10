# OpenEdge Configuration Interpretation

- [Statement](#statement)
- [Master Configuration](#master-configuration)
- [Application Configuration](#application-configuration)
- [openedge-agent Configuration](#openedge-agent-configuration)
- [openedge-hub Configuration](#openedge-hub-configuration)
- [openedge-function-manager Configuration](#openedge-function-manager-configuration)
- [openedge-function-python27、openedge-function-python36 Configuration](#openedge-function-python27-configuration-openedge-function-python36-configuration)
- [openedge-remote-mqtt Configuration](#openedge-remote-mqtt-configuration)

## Statement

Supported units:

- Size unit: b(byte), k(kilobyte), m(megabyte), g(gigabyte)
- Time unit: s(second), m(minute), h(hour)

Configuration examples can be found in the `example` directory of this project.

## Master Configuration

The Master configuration and application configuration are separated. The default configuration file is `etc/openedge/openedge.yml` in the working directory. The configuration is interpreted as follows:

```yaml
mode: The default value is `docker`, running mode of services. **docker** container mode or **native** process mode
grace: The default value is `30s`, the timeout for waiting services to gracefully exit.
server: API Server configuration of Master.
  address: The default value can be read from environment variable `OPENEDGE_MASTER_API`, address of API Server.
  timeout: The default value is `30s`, timeout of API Server.
logger: Logger configuration
  path: The default is `empty` (none configuration), that is, it does not write to the file. If the path is specified, it writes to the file.
  level: The default value is `info`, log level, support `debug`、`info`、`warn` and `error`.
  format: The default value is `text`, log print format, support `text` and `json`.
  age:
    max: The default value is `15`, means maximum number of days the log file is kept.
  size:
    max: The default value is `50`, log file size limit, default unit is `MB`.
  backup:
    max: The default value is `15`, the maximum number of log files to keep.
```

## Application Configuration

The default configuration file for the application configuration is `var/db/openedge/application.yml` in the working directory. The configuration is interpreted as follows:

```yaml
version: Application version
services: Service list configuration
  - name: [MUST] Service name, must be unique in the service list
    image: [MUST] Service entry. In the docker container mode, which means the address of image. In the native process mode indicates where the service program package is located.
    replica: The default value is 0, the number of service copies, indicating the number of service instances started. Usually the service only needs to start one. The function runtime service is generally set to 0, not started by the Master, but is dynamically started by the function manager service.
    mounts: Storage volume mapping list
      - name: [MUST] The volume name, corresponding to one of the storage volume lists
        path: [MUST] The path mapped by the volume in the container
        readonly: The default value is false, whether the storage volume is read-only
    ports: Ports exposed in docker container mode, for example
      - 0.0.0.0:1883:1883
      - 0.0.0.0:1884:1884/tcp
      - 8080:8080/tcp
      - 9884:8884
    devices: Device mapping in docker container mode, for example
      - /dev/video0
      - /dev/sda:/dev/xvdc:r
    args: Service instance startup arguments, for example
      - '-c'
      - 'conf/conf.yml'
    env: Environment variables of service instance, for example
      version: v1
    restart: Service restart policy configuration
      retry:
        max: The default is `empty`(none configuration), which means always retry. If not, which means the maximum number of service restarts.
      policy: The default value is `always`, restart policy, support `no`, `always` and `on-failure`. And `no` means none restart, `always` means always restart, `on-failure` means restart the service if it exits abnormally.
      backoff:
        min: The default value is `1s`, minimum interval of restart.
        max: The default value is `5m`, maximum interval of restart.
        factor: The default value is `2`, factor of interval increase.
    resources: Service instance resource limit configuration in docker container mode
      cpu:
        cpus: The percentage of CPU available of the service instance, for example `1.5`, means that `1.5` CPU cores can be used.
        setcpus: The CPU core available for the service instance, for example `0-2`, means that `0` to `2` CPU cores can be used; `0` means that the 0th CPU core can be used; `1`, which means the 1st CPU core can be used.
      memory:
        limit: The available memory of the service, for example `500m`, means that 500 megabytes of memory can be used.
        swap: The swap space available to the service, for example `1g`, means that 1G of memory can be used.
      pids:
        limit: Number of processes the service can create.
volumes: Storage volume list
  - name: [MUST] The volume name, must be unique in the list of storage volumes
    path: [MUST] The path of the storage volume on the host, relative to the working directory of the Master
```

## openedge-agent Configuration

```yaml
remote:
  mqtt: MQTT channel configuration
    clientid: [MUST] The Client ID, must be the id of cloud core device.
    address: [MUST] The endpoint address for client to connect with cloud management suit, must use ssl endpoint.
    username: [MUST] The client username, must be the username of cloud core device.
    ca: [MUST] The CA path for client to connect with cloud management suit.
    key: [MUST] The private key path for client to connect with cloud management suit.
    cert: [MUST] The public key path for client to connect with cloud management suit.
    timeout: The default value is `30s`, means timeout of the client connects to cloud.
    interval: The default value is `1m`, means maximum interval of client reconnection, doubled from 500 microseconds to maximum.
    keepalive: The default value is `1m`, means keep alive time between the client and cloud after connection has been established.
    cleansession: The default value is `false`, , means whether keep session in cloud after client disconnected.
    validatesubs: The default value is `false`, means whether the client checks the subscription result. If it is true, client exits and return errors when subscription is failure.
    buffersize: The default value is `10`, means the size of the memory queue sent by the client to the cloud management suit. If found exception, the client will exit and lose message.
  http: HTTPS channel configuration
    address: This address is automatically inferred based on the address of the MQTT channel. No configuration required
    timeout: The default value is `30s`, connection timeout period
  report: Agent report configuration.
    url: The report URL. No configuration required
    topic: The template of report topic. No configuration required
    interval: The default value is `20s`, interval of reporting.
  desire: Agent desire configuration.
    topic: The template of desire topic. No configuration required
```

## openedge-hub Configuration

```yaml
listen: [MUST] Listening address, for example
  - tcp://0.0.0.0:1883
  - ssl://0.0.0.0:1884
  - ws://:8080/mqtt
  - wss://:8884/mqtt
certificate: SSL/TLS certificate authentication configuration, if `ssl` or `wss` is enabled, it must be configured.
  ca: Server CA certificate path
  key: Server private key path
  cert: Server public key path
principals: ACL configuration. If not configured, client cannot connect to this Hub, support username/password and certificate authentication.
  - username: Username for client non-ssl connection
    password: Password for client connection
    permissions:
      - action: Operation type of permission. `pub` means publish permission, `sub` means subscription permission.
        permit: List of topics allowed by the operation type, support `+` and `#` wildcards.
  - username: Username for client ssl connection
    permissions:
      - action: Operation type of permission. `pub` means publish permission, `sub` means subscription permission.
        permit: List of topics allowed by the operation type, support `+` and `#` wildcards.
subscriptions: Topic routing configuration
  - source:
      topic: subscribe topic
      qos: QoS of topic
    target:
      topic: publish topic
      qos: QoS of topic
message: MQTT message related configuration
  length:
    max: The default value is `32k`, which means maximum message length that can be allowed to be transmitted. The maximum can be set to 268,435,455 Byte(about 256MB).
  ingress: Message receive configuration
    qos0:
      buffer:
        size: The default value is `10000`, means the number of messages that can be cached in memory with QoS0. Increasing the cache can improve the performance of message reception. If the device loses power, it will directly discard the message with QoS0.
    qos1:
      buffer:
        size:  The default value is `100`, means the message cache size of waiting for persistent with QoS1. Increasing the cache can improve the performance of message reception, but the potential risk is that the service will exit abnormally(such as device power failure), it will lose the cached message, and will not reply(puback). The service exits normally and waits for the cached message to be processed without losing data.
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
      size:  The default value is `10000`, means the size of the cache queue for the serial number of the message that was acknowledged(ack). For example, three messages with QoS1 and serial numbers 1, 2, and 3 are sent to the client in batches. The client confirms the messages of sequence numbers 1 and 3. At this time, sequence number 1 will be queued and persisted. Although sequence number 3 has been confirmed, it still has to wait for the serial number 2 to be confirmed before entering the column. This design can ensure that the message can be recovered from the persistent serial number after the service restarts abnormally, ensuring that the message is not lost, but the message retransmission will occur, and therefore the message with QoS 2 is not supported.
    batch:
      max:  The default value is `100`, means the maximum number of batches of message serial numbers can be insert into the database.
logger: Logger configuration
  path: The default is `empty` (none configuration), that is, it does not write to the file. If the path is specified, it writes to the file.
  level: The default value is `info`, log level, support `debug`、`info`、`warn` and `error`.
  format: The default value is `text`, log print format, support `text` and `json`.
  age:
    max: The default value is `15`, means maximum number of days the log file is kept.
  size:
    max: The default value is `50`, log file size limit, default unit is `MB`.
  backup:
    max: The default value is `15`, the maximum number of log files to keep.
status: Service status configuration
  logging:
    enable: The default value is `false`, means whether to print openedge status information.
    interval: The default value is `60s`, means interval of printing openedge status information.
storage: Database storage configuration
  dir: The default value is `var/db/openedge/data`, means database storage directory.
shutdown: Service exit configuration
  timeout: The default value is `10m`, means timeout of service exit.
```

## openedge-function-manager Configuration

```yaml
hub:
  clientid: [MUST] The Client ID for the client to connect with the local Hub.
  address: [MUST] The endpoint address for the client to connect with the local Hub.
  username: The username for the client to connect with the local hub.
  password: The password for the client to connect with the local hub.
  ca: The CA path for the client to connect with the local hub.
  key: The private key path for the client to connect with the local hub.
  cert: The public key path for the client to connect with the local hub.
  timeout: The default value is `30s`, means timeout of the client connection with the local hub.
  interval: The default value is `1m`, means maximum interval of client reconnection, doubled from 500 microseconds to maximum.
  keepalive: The default value is `1m`, means keep alive time between the client and the local hub after connection has been established.
  cleansession: The default value is `false`, , means whether keep session in the local Hub after client disconnected.
  validatesubs: The default value is `false`, means whether the client checks the subscription result. If it is true, client exits and return errors when subscription is failure.
  buffersize: The default value is `10`, means the size of the memory queue sent by the client to the local Hub. If found exception, the client will exit and lose messages.
rules: Router rules configuration
  - clientid: [MUST] The Client ID for client to connect with the local Hub
    subscribe:
      topic: [MUST] The message topic subscribed from the local Hub.
      qos: The default value is `0`, the message QoS subscribed from the local Hub.
    function:
      name: [MUST] The name of the function that processes the message.
    publish:
      topic: [MUST] The message topic published to the local Hub.
      qos: The default value is `0`, means the message QoS published to the local Hub.
functions:
  - name: [MUST] The function name, must be unique in the function list.
    service: [MUST] The service name which provides the function runtime instance.
    instance: function instance configuration
      min: The default value is `0`, means the minimum number of function instance. And the minimum configuration allowed to be set is `0`, the maximum configuration allowed to be set is `100`.
      max: The default value is `1`, means the maximum number of function instance. And the minimum configuration allowed to be set is `1`, the maximum configuration allowed to be set is `100`.
      idletime: The default value is `10m`, maximum idle time of function instance.
      evicttime: The default value is `1m`, interval time between two evict operations.
      message:
        length:
          max: The default value is `4m`, means the maximum message length allowed for function instances to be received and publish.
    backoff:
      max: The default value is `1m`, the maximum reconnection interval of the client connection function instance
    timeout: The default value is `30s`, Client connection function instance timeout
```

## openedge-function-python27、openedge-function-python36 Configuration

```yaml
# the configurations of the two modules are the same, so we can follow this sample below
server: GRPC Server configuration; Do not configure if the instances of this service are managed by openedge-function-manager
  address: GRPC Server address, <host>:<port>
  workers:
    max: The default value is the number of CPU core multiplied by 5, the maximum capacity of the thread pool
  concurrent:
    max: The default value is `empty`, means no limit, the maximum number of concurrent connections
  message:
    length:
      max: The default value is `4m`, the maximum message length allowed for function instances to receive and send
  ca: Server CA certificate path
  key: Server private key path
  cert: Server public key path
functions: function list
  - name: [MUST] The function name, must be unique in the function list.
    handler: [MUST] The function of Python code to handle message, for example, 'sayhi.handler'
    codedir: [MUST] The path of Python code
logger: Logger configuration
  path: The default is `empty` (none configuration), that is, it does not write to the file. If the path is specified, it writes to the file.
  level: The default value is `info`, log level, support `debug`、`info`、`warn` and `error`.
  format: The default value is `text`, log print format, support `text` and `json`.
  age:
    max: The default value is `15`, means maximum number of days the log file is kept.
  size:
    max: The default value is `50`, log file size limit, default unit is `MB`.
  backup:
    max: The default value is `15`, the maximum number of log files to keep.
```

## openedge-remote-mqtt Configuration

```yaml
hub:
  clientid: [MUST] The Client ID for the client to connect with the local Hub.
  address: [MUST] The endpoint address for the client to connect with the local Hub.
  username: The username for the client to connect with the local hub.
  password: The password for the client to connect with the local hub.
  ca: The CA path for the client to connect with the local hub.
  key: The private key path for the client to connect with the local hub.
  cert: The public key path for the client to connect with the local hub.
  timeout: The default value is `30s`, means timeout of the client connection with the local hub.
  interval: The default value is `1m`, means maximum interval of client reconnection, doubled from 500 microseconds to maximum.
  keepalive: The default value is `1m`, means keep alive time between the client and the local hub after connection has been established.
  cleansession: The default value is `false`, , means whether keep session in the local Hub after client disconnected.
  validatesubs: The default value is `false`, means whether the client checks the subscription result. If it is true, client exits and return errors when subscription is failure.
  buffersize: The default value is `10`, means the size of the memory queue sent by the client to the local Hub. If found exception, the client will exit and lose messages.
rules: Message routing rules configuration
  - hub:
      clientid: The client ID for the client to connect with the local Hub.
      subscriptions: The topics subscribed by client from Hub, for example
        - topic: say
          qos: 1
        - topic: hi
          qos: 0
    remote:
      name: [MUST] The remote name, must be one of the remote list
      clientid: The client ID for the client to connect with remote Hub.
      subscriptions: The topics subscribed by client from remote Hub, for example
        - topic: remote/say
          qos: 0
        - topic: remote/hi
          qos: 0
remotes: The remote list
  - name: [MUST] The remote name, must be unique in this list.
    clientid: The client ID for the client to connect with the remote Hub.
    address: [MUST] The address for the client connect with the remote Hub.
    username: The username for the client connect with the remote Hub.
    password: The password for the client connect with the remote Hub.
    ca: The CA path for the client connect with the remote Hub.
    key: The private key path for the client connect with the remote Hub.
    cert: The public key path for the client connect with the remote Hub.
    timeout: The default value is `30s`, means timeout of the client connect to the remote Hub.
    interval: The default value is `1m`, means maximum interval of client reconnection, doubled from 500 microseconds to maximum.
    keepalive: The default value is `1m`, means keep alive time between the client and the local hub after connection has been established.
    cleansession: The default value is `false`, , means whether keep session in the local Hub after client disconnected.
    validatesubs: The default value is `false`, means whether the client checks the subscription result. If it is true, client exits and return errors when subscription is failure.
    buffersize: The default value is `10`, means the size of the memory queue sent by the client to the local Hub. If found exception, the client will exit and lose messages.
logger: Logger configuration
  path: The default is `empty` (none configuration), that is, it does not write to the file. If the path is specified, it writes to the file.
  level: The default value is `info`, log level, support `debug`、`info`、`warn` and `error`.
  format: The default value is `text`, log print format, support `text` and `json`.
  age:
    max: The default value is `15`, means maximum number of days the log file is kept.
  size:
    max: The default value is `50`, log file size limit, default unit is `MB`.
  backup:
    max: The default value is `15`, the maximum number of log files to keep.
```

## openedge-timer Configuration

```yaml
hub: Hub configuration
  address: The address for the client to connect with the Hub.
  username: The username for the client to connect with the Hub.
  password: The password for the client to connect with the Hub.
  clientid: The client id for the client to connect with the Hub.
timer: timer configuration
  interval: Timing interval
publish:
  topic: The message topic published to the Hub.
  payload: The payload data, for example
    id: 1
logger: Logger configuration
  path: The default is `empty` (none configuration), that is, it does not write to the file. If the path is specified, it writes to the file.
  level: The default value is `info`, log level, support `debug`、`info`、`warn` and `error`.
```
