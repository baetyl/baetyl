# Pre-release 0.1.3(2019-05-10)

## Features

- Supports the custom status information of the service instance and collects more system status information. For details, please refer to: [https://github.com/baidu/openedge/blob/master/doc/zh-cn/overview/OpenEdge-design .md#system-inspect](https://github.com/baidu/openedge/blob/master/doc/zh-cn/overview/OpenEdge-design.md#system-inspect)
- When the openedge starts, the old instance will be cleaned up (the residual instance is usually caused by the abnormal exit of the openedge)
- Added Python 3.6 version of the function runtime, using Ubuntu16.04 as the base image
- Docker container mode supports runtime and args configuration
Added a simple timer module openedge-timer

## Bug Fixs

- When the function instance pool destroys the function instance, be sure to stop the function instance
- The openedge stop command waits for the openingge to stop running and exits, ensuring that the pid file is cleaned up
- The hub module posts a message waiting for the ack to time out and quickly resend the message.
- Resolved an issue where the argument to atomic.addUint64() did not follow the 64-bit alignment causing the exception to exit. Reference: [https://github.com/golang/go/issues/23345](https://github.com/golang/go/issues/23345)

## Others(include release engineering)

- Release OpenEdge 2019 Roadmap
- Release of the OpenEdge Community Participant Convention