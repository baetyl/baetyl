/*
Title: Baetyl 0.1.4
Sort: 20
*/

# Pre-release 0.1.4(2019-07-05)

## New features

- [#251](https://github.com/baidu/openedge/issues/251) add node85 function runtime
- [#260](https://github.com/baidu/openedge/issues/260) collect network ip address and MAC information
- [#263](https://github.com/baidu/openedge/issues/263) optimize app reload logic in master, keep service running of its config not changed
- [#264](https://github.com/baidu/openedge/issues/264) optimize volume clean logic and move it from master to agent module, will remove all volumes not in app's volumes list
- [#266](https://github.com/baidu/openedge/issues/266) stats the cpu and memory of the service instances

## Bugfixes

- [#246](https://github.com/baidu/openedge/issues/246) change the interval of stats report of agent module from 1m to 20s

## Others

- [#269](https://github.com/baidu/openedge/issues/269) [#273](https://github.com/baidu/openedge/issues/273) [#280](https://github.com/baidu/openedge/issues/280) update makefile, support selected deploy
