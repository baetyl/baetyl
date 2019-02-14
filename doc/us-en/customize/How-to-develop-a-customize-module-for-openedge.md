# How to develop a customize module for OpenEdge

Read the Development Compilation Guide before developing and integrating custom modules to understand OpenEdge's build environment requirements.

Custom modules do not limit the development language. Understand these conventions below to integrate custom modules better and faster.

## Directory Convention

### Docker Container Mode

The working directory in the container is: /

The configuration path in the container is: /etc/openedge/module.yml

The resource file directory in the container is: /var/db/openedge/module/<module name>

The persistent data output directory in the container is: /var/db/openedge/volume/<module name>

The persistent log output directory in the container is: /var/log/openedge/<module name>

The docker volumes mapping is as follows:

> - <openedge_host_work_dir>/var/db/openedge/module/<module_name>/module.yml:/etc/openedge/module.yml
> - <openedge_host_work_dir>/var/db/openedge/module/<module_name>:/var/db/openedge/module/<module_name>
> - <openedge_host_work_dir>/var/db/openedge/volume/<module_name>:/var/db/openedge/volume/<module_name>
> - <openedge_host_work_dir>/var/log/openedge/<module_name>:/var/log/openedge/<module_name>

_**NOTE**: If the data needs to be persisted on the device (host), such as the database and log, it must be saved in the persistent directory specified above, otherwise after the container destruction data will be lost._

### Native Process Mode

The working directory of the module is the same as the working directory of the OpenEdge master program.

The configuration path of the module is: <openedge_host_work_dir>/var/db/openedge/module/<module name>/module.yml

The module's resource file directory is: <openedge_host_work_dir>/var/db/openedge/module/<module name>

The module's data output directory is: <openedge_host_work_dir>/var/db/openedge/volume/<module name>

The module's log output directory is: <openedge_host_work_dir>/var/log/openedge/<module name>

## Configuration Convention

The module supports loading the yaml format configuration from the file, reading /etc/openedge/module.yml in Docker container mode, and reading <openedge_host_work_dir>/var/db/openedge/module/<module name>/module.yml in native process mode.

It also supports getting the configuration from the input parameters, which can be a string in json format. such as:

    modules:
      - name: 'my_module'
        entry: 'my_module_docker_image'
        params:
          - '-c'
          - '{"name":"my_module","address":"127.0.0.1:1234",...}'


## Start/Stop Convention

The module is started as a process independently by master with module's configuration, and the module should listen to the SIGTERM signal to gracefully exit when stopped by master. A simple golang module implementation can refer to [MQTT Remote Module](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/openedge-remote-mqtt).

## Module SDK

If the module is developed using golang, you can use the module SDK provided in openedge, located at github.com/baidu/openedge/module.

[mqtt.Dispatcher](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/module/mqtt/dispatcher.go) can be used to subscribe to the mqtt server and support automatic reconnection. The mqtt dispatcher does not support message persistence. The message persistence should be handled by the mqtt hub. If the message subscribed by the mqtt dispatcher is 1, the message needs to be replied to ack after the message is processed, otherwise the hub will resend the message. [remote mqtt module reference](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/openedge-remote-mqtt/main.go)

[runtime.Server](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/module/function/runtime/server.go) encapsulates the grpc server and function invoke logic, which is convenient for developers to implement the function runtime of message processing. [python2.7 runtime reference]
(https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/openedge-function-runtime-python27/openedge_function_runtime_python27.py)

[master.Client](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/module/master/client.go) can be used to call the master program's API to start or stop the temporary module. The account username is the resident module name and the password is obtained from the environment variable using ```module.GetEnv(module.EnvOpenEdgeModuleToken)```.

[module.Load](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/module/module.go) can be used to load the configuration of the module, support to read the configuration from file in yaml format and from process arguments in json format.

[module.Wait](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/module/module.go) can be used to wait the SIGTERM signal for the module to exit.

[logger](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/module/logger/logger.go) can be used to log.
