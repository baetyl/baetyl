# Pre-release 0.1.2(2019-04-04)

## Features

- Separate the agent module from the master and report the status periodically
- Introduce volume to abstract resources in configuration and support existing images, such as mosquitto from hub.docker.com
- Publish the command line and support background startup
- Uniform configuration of the two modes, such as create a separate working directory for each service in native process mode
- Introduce service replace module and support to start multiple instance
- Support device mapping in docker container mode

## Bug fixes

- Add `openedge.sock` clean logic
- Upgrade openedge-hub, change auth logic of password and tls
- Upgrade openedge-function-x, add retry logic and remove keep order logic

## Others(include release engineering)

- Rich test example support, such as for hub module, provide mosquitto configuration
- All documents in English

# Pre-release 0.1.1(2018-12-28)

## Features

- optimize MQTT Remote module, support multiple remotes and message router
- add validatesubs config to check MQTT client subscribe result or not

## Bug fixes

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
- provide some official modules, such as hub, function(and python2.7 runtime), MQTT remote, etc.

## Bug Fixes

- N/A
