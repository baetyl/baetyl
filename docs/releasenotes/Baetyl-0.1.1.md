/*
Title: Baetyl 0.1.1
Sort: 50
*/

# Pre-release 0.1.1(2018-12-28)

## New features

- optimize MQTT Remote module, support multiple remotes and message router
- add validate subscription config to check MQTT client subscribe result or not

## Bugfixes

- isolate module directory in docker mode
- remove network scope filter to support old docker

## Others(include release engineering)

- rich build, test and release scripts and documents
- check in vendor to solve all build issues
- refactor code and format all messages
- replace shell with makefile
- update gomqtt
- add travis CI