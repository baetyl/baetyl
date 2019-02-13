# Pre-release 0.1.1(2018-12-28)

## Features

> + optimize MQTT Remote Module, support multiple remotes and message router
> + add validatesubs config to check MQTT client subscribe result or not

## Bug fixes

> + isolate module directory in docker mode
> + remove network scope filter to support old docker

## Others(include release engineering)

> + rich build, test and release scripts and documents
> + checkin vendor to solve all build issues
> + refactor code and format all messages
> + replace shell with makefile
> + update gomqtt
> + add travis CI

# Pre-release 0.1.0(2018-12-05)

Initial open source.

## Features

> + support module management after modularization
> + support two modes: docker container and native process
> + support resource constraints in docker container mode(cpu, memory, etc)
> + provide some official modules, such as hub, function(and python2.7 runtime), MQTT remote, etc.

## Bug Fixes

> + N/A
