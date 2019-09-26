/*
Title: Baetyl 0.1.3
Sort: 30
*/

# Pre-release 0.1.3(2019-05-10)

## New features

- [#199](https://github.com/baidu/openedge/issues/199) Supports the custom status information of the service instance and collects more system status information. For details, please refer to: [https://github.com/baidu/openedge/blob/master/doc/zh-cn/overview/Baetyl-design .md#system-inspect](https://github.com/baidu/openedge/blob/master/doc/zh-cn/overview/Baetyl-design.md#system-inspect)
- [#209](https://github.com/baidu/openedge/issues/209) When the openedge starts, the old instance will be cleaned up (the residual instance is usually caused by the abnormal exit of the openedge)
- [#211](https://github.com/baidu/openedge/issues/211) Added Python 3.6 version of the function runtime, using Ubuntu16.04 as the base image
- [#222](https://github.com/baidu/openedge/issues/222) Docker container mode supports runtime and args configuration
- Added a simple timer module openedge-timer

## Bugfixes

- [#201](https://github.com/baidu/openedge/issues/201) When the function instance pool destroys the function instance, be sure to stop the function instance
- [#208](https://github.com/baidu/openedge/issues/208) The openedge stop command waits for the openingge to stop running and exits, ensuring that the pid file is cleaned up
- [#234](https://github.com/baidu/openedge/issues/234) The hub module posts a message waiting for the ack to time out and quickly resend the message.
- atomic.addUint64() panics if the pointer to its argument is not 64byte aligned. ref: https://github.com/golang/go/issues/23345

## Others(include release engineering)

- [#230](https://github.com/baidu/openedge/issues/230) Release Baetyl 2019 Roadmap
- [#228](https://github.com/baidu/openedge/issues/228) Release of the Baetyl Community Participant Convention