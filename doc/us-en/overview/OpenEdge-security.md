# OpenEdge Security

For security reasons, OpenEdge supports full platform security certificate authentication mode. Generally, all devices, applications and services connected to OpenEdge should be authenticated through certificate. For the various modules of OpenEdge, the security design and authentication policy are slightly different, as follows:

- For Hub module, it mainly provides device connect abilities, currently supports `TCP`, `SSL`(TCP + SSL), `WS`(Websocket) and `WSS`(Websocket + SSL) 4 connection methods.
   - Among them, for `SSL` connection method, OpenEdge supports one-way and two-way authentication of certificate.
- For MQTT Remote module, highly recommended users to use (SSL)certificate authentication.
- For Agent module, OpenEdge enforces users to use HTTPS protocol to ensure the security of information reporting.
- For Agent module, OpenEdge also enforces users to use HTTPS protocol to ensure the security of configuration delivery.

In general, for different modules and services, OpenEdge provides multiple ways to ensure the security of information interaction.