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