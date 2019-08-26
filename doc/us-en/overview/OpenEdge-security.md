# Baetyl Security

For security reasons, Baetyl supports full platform security certificate authentication mode. Generally, all devices, applications and services connected to Baetyl should be authenticated through certificate. For the various modules of Baetyl, the security design and authentication policy are slightly different, as follows:

- For Hub module, it mainly provides device connect abilities, currently supports `TCP`, `SSL`(TCP + SSL), `WS`(Websocket) and `WSS`(Websocket + SSL) 4 connection methods.
   - Among them, for `SSL` connection method, Baetyl supports one-way and two-way authentication of certificate.
- For MQTT Remote module, highly recommended users to use (SSL)certificate authentication.
- For Agent module, Baetyl enforces users to use HTTPS protocol to ensure the security of information reporting.
- For Agent module, Baetyl also enforces users to use HTTPS protocol to ensure the security of configuration delivery.

In general, for different modules and services, Baetyl provides multiple ways to ensure the security of information interaction.