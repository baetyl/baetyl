# OpenEdge security

For security reasons, OpenEdge supports full platform security certificate authentication mode. Generally, all devices, applications and services connected to OpenEdge should be authenticated through certificate. For the various modules of OpenEdge, the security design and authentication policy are slightly different, as follows:

> + For Hub module, it mainly provides device connect abilities, currently supports `tcp`, `ssl`(tcp ssl), `ws`(websocket) and `wss`(websocket ssl) 4 connection modes.
>   - Among them, for `ssl` connect mode, OpenEdge supports one-way and two-way authentication of certificate.
> + For MQTT Remote Module, in the case of using BIE cloud console management suite(recommended), only certificate authentication configuration is supported. If users do not use BIE cloud console to delivery configuration, the username/password authentication method also can be supported.
> + For device hardware information reporting service, OpenEdge enforces users to use HTTPS protocol to ensure the security of information reporting.
> + For configuration delivered service, OpenEdge also enforces users to use HTTPS protocol to ensure the security of the configuration.

In general, for different modules and services, OpenEdge provides multiple ways to ensure the security of information interaction.