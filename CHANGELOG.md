# Pre-release 0.1.3(2019-05-10)

## New features

- [#199](https://github.com/baidu/openedge/issues/199) support to report custom stats for service instances, collect more system stats
- [#209](https://github.com/baidu/openedge/issues/209) clean up old instances if exists when openedge starts
- [#211](https://github.com/baidu/openedge/issues/211) add new function runtime Python3.6, and use ubuntu 16.04 as the base image
- [#222](https://github.com/baidu/openedge/issues/222) support runtime for docker container mode
- add a simple timer module: openedge-timer

## Bugfixes

- [#201](https://github.com/baidu/openedge/issues/201) stop the function instance if it cannot be returned to pool
- [#208](https://github.com/baidu/openedge/issues/208) wait openedge to stop, clean pid file if need
- [#234](https://github.com/baidu/openedge/issues/234) quickly republish message if ack timeout in hub module
- atomic.addUint64() panics if the pointer to its argument is not 64byte aligned. ref: https://github.com/golang/go/issues/23345

## Others

- [#230](https://github.com/baidu/openedge/issues/230) Add 2019 roadmap
- [#228](https://github.com/baidu/openedge/issues/228) Add code of conduct

# Pre-release 0.1.2(2019-04-04)

## New features

- [#20](https://github.com/baidu/openedge/issues/20) Separate the agent module from the master and report the status periodically
- [#120](https://github.com/baidu/openedge/issues/120) Introduce volume to abstract resources in configuration and support existing images, such as mosquitto from hub.docker.com
- [#122](https://github.com/baidu/openedge/issues/122) Publish the command line and support background startup
- Uniform configuration of the two modes, such as create a separate working directory for each service in native process mode
- [#123](https://github.com/baidu/openedge/issues/123) Introduce service replace module and support to start multiple instance
- [#142](https://github.com/baidu/openedge/issues/142) Support device mapping in docker container mode

## Bugfixes

- [#92](https://github.com/baidu/openedge/issues/92) [#81](https://github.com/baidu/openedge/issues/81) Add `openedge.sock` clean logic
- [#135](https://github.com/baidu/openedge/issues/135) [#88](https://github.com/baidu/openedge/issues/88) Upgrade openedge-hub, change auth logic of password and tls
- [#127](https://github.com/baidu/openedge/issues/127) Upgrade openedge-function-x, add retry logic and remove keep order logic

## Others(include release engineering)

- Rich test example support, such as for hub module, provide mosquitto configuration
- All documents in English

# Pre-release 0.1.1(2018-12-28)

## New features

- optimize MQTT Remote module, support multiple remotes and message router
- add validatesubs config to check MQTT client subscribe result or not

## Bugfixes

- isolate module directory in docker mode
- remove network scope filter to support old docker

## Others(include release engineering)

- rich build, test and release scripts and documents
- checkin vendor to solve all build issues
- refactor code and format all messages
- replace shell with makefile
- update gomqtt
- add travis CI

# Pre-release 0.1.0(2018-12-05)

Initial open source.

## Features

- support module management after modularization
- support two modes: docker container and native process
- support resource constraints in docker container mode(cpu, memory, etc)
- provide some official modules, such as hub, function(and Python2.7 runtime), MQTT remote, etc.

## Bugfixes

- N/A
