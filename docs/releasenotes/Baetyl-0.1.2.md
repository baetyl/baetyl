/*
Title: Baetyl 0.1.2
Sort: 40
*/

# Pre-release 0.1.2(2019-04-04)

## New features

- Separate the agent module from the master and report the status periodically
- Introduce volume to abstract resources in configuration and support existing images, such as mosquitto from hub.docker.com
- Publish the command line and support background startup
- Uniform configuration of the two modes, such as create a separate working directory for each service in native process mode
- Introduce service replace module and support to start multiple instance
- Support device mapping in docker container mode

## Bugfixes

- Add `openedge.sock` clean logic
- Upgrade openedge-hub, change auth logic of password and tls
- Upgrade openedge-function-x, add retry logic and remove keep order logic

## Others(include release engineering)

- Rich test example support, such as for hub module, provide mosquitto configuration
- All documents in English