/*
Title: Baetyl 0.1.5
Sort: 10
*/

# Pre-release 0.1.5(2019-08-15)

## New features

- [#302](https://github.com/baidu/openedge/pull/302) support master OTA, new OTA logic for master and app
- [#305](https://github.com/baidu/openedge/pull/305) support config template in golang sdk
- [#300](https://github.com/baidu/openedge/pull/300) [#301](https://github.com/baidu/openedge/pull/301) support systemd in deb installation
- [#298](https://github.com/baidu/openedge/pull/298) collect core num of CPU of host
- [#297](https://github.com/baidu/openedge/pull/297) support sock config and file mount
- [#290](https://github.com/baidu/openedge/pull/290) support deb build and publish
- [#289](https://github.com/baidu/openedge/pull/289) run openedge in the foreground (openedge should be managed by daemon tools, for example, systemd)
- [#293](https://github.com/baidu/openedge/pull/293) add new function runtime openedge-function-python36 with opencv 4.1.0 to handle images or AI inference results

## Bugfixes

- [#303](https://github.com/baidu/openedge/pull/303) fix "address already in use" issue of openedge.sock
- [#292](https://github.com/baidu/openedge/pull/292) check service list in app config in agent module

## Others

- N/A