# Message handling with Local Function Module

**Statement**

> + The operating system as mentioned in this document is Darwin.
> + The MQTT client toolkit as mentioned in this document is [MQTTBOX](../Resources-download.md#mqttbox-download).
> + The docker image used in this document is compiled from the OpenEdge source code. More detailed contents please refer to [Build OpenEdge from source](../setup/Build-OpenEdge-from-Source.md)

Different from the Local Hub Module to transfer message among devices(mqtt clients), this document describes the message handling with Local Function Module(also include Local Hub Module and Pyton27 Runtime Module). In the document, Local Hub Module is used to establish connection between OpenEdge and mqtt client, Python27 Runtime Module is used to hanle MQTT messages, and the Local Function Module is used to combine Local Hub Module with Python27 Runtime Module with message context.


## Workflow



## Message Handling Test