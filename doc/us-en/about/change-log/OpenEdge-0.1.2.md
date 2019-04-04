# Pre-release 0.1.2(2019-04-04)

## Features

- [Issue 20](https://github.com/baidu/openedge/issues/20) Separate the agent module from the master and report the status periodically
- [Issue 120](https://github.com/baidu/openedge/issues/120)  Introduce volume to abstract resources in configuration and support existing images, such as mosquitto from hub.docker.com
- [Issue 122](https://github.com/baidu/openedge/issues/122)  Publish the command line and support background startup
- Uniform configuration of the two modes, such as create a separate working directory for each service in native process mode
- [Issue 123](https://github.com/baidu/openedge/issues/123) Introduce service replace module and support to start multiple instance
- [Issue 142](https://github.com/baidu/openedge/issues/142) Support device mapping in docker container mode

## Bug fixes

- [Issue 81](https://github.com/baidu/openedge/issues/81) [Issue 92](https://github.com/baidu/openedge/issues/92) Add openedge.sock clean logic
- [Issue 88](https://github.com/baidu/openedge/issues/88) [Issue 135](https://github.com/baidu/openedge/issues/135) Upgrade openedge-hub, change auth logic of password and tls
- [Issue 127](https://github.com/baidu/openedge/issues/127) Upgrade openedge-function-x, add retry logic and remove keep order logic

## Others(include release engineering)

- Rich test example support, such as for hub module, provide mosquitto configuration
- [Issue 61](https://github.com/baidu/openedge/issues/61) All documents in English