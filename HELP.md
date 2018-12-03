# 声明

本文档用于记录用户可能遇到和提出的问题。

+ **Q:** 为什么我下载了智能边缘运行包，解压后直接启动显示无法启动？

![智能边缘运行包在错误系统平台上的运行示意图](http://agroup.baidu.com:8964/static/90/6ed662af2fe692f0026378f1ba374dcc13391e.png?filename=exec+format+error.png)

  **A:** 导致该问题的原因可能有多种情况，请首先观察智能边缘运行包执行终端有无启动失败的异常日志提示，若有异常日志提示，则可能是由于配置文件的配置有误导致（智能边缘启动时候会对配置文件所有配置项进行合法性检查、校验），请查阅README确认正确配置信息，修正后再次启动即可；若无异常日志提示，则可能是由于下载的智能边缘运行包平台有误（智能边缘运行包与系统平台和CPU架构强相关，请确认部署机器平台后针对性[下载](https://cloudtest.baidu.com/doc/BIE/BIEdownload.html#.72.C5.4B.29.68.F9.F2.6C.B5.23.B6.09.7F.15.D5.D6)）。

----------

+ **Q:** 为什么我启动智能边缘运行包后，用配置文件配置的用户名和密码无法连接智能边缘？

  **A:** 首先，智能边缘对配置文件中关于设备权限主题的配置中要求用户名的合法字符范围是[0-9A-Za-z_-]，长度不超过32位，且密码要求统一用SHA256编码加密处理，权限主题约束规则依据MQTT协议订阅主题过滤器约束规则。另外，在使用客户端与智能边缘进行连接时，要求连接密码使用SHA256进行哈希之前的明文形式。

----------

+ **Q:** 为什么我通过客户端正常连接智能边缘后，无法订阅某个主题或无法向某个主题发布消息？

  **A:** 这个问题也是可能由多种情况导致，其一是由于对MQTT协议了解不够导致，MQTT协议要求可订阅或发布的主题规则可参考具体[MQTT协议文档](http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.html)；其二是由于智能边缘与连接设备间的通信和数据处理均依赖主题，且要求通信所用主题必须在Principals中进行配置，只有配置相应权限的主题才能够正常使用；其三是由于客户端向智能边缘发布的消息内容过多，超过了智能边缘所允许的范围（消息Payload允许的最大长度）导致。

----------

+ **Q:** 为什么我在subscriptions中配置了函数，启动智能边缘运行包后报错？

  **A:** 这是由于在subscriptions中配置了函数后，需要在functions中配置对应的函数信息，如runtime、handler等。

----------

+ **Q:** 为什么我在subscriptions中配置了Remote远程服务，启动智能边缘运行包后报错？

  **A:** 这是由于在subscriptions中配置了Remote远程服务后，需要在remotes中配置对应的远程服务信息。

----------

+ **Q:** 为什么我在subscriptions中配置了主题路由转发策略，启动智能边缘运行包进行数据处理后没有收到处理后的消息？

  **A:** 如配置路由转发策略subscription{source: “test/from”, target: “test/to”}，即表明将主题”test/from”的消息以新主题”test/to”发布到openedge内部broker，client端需订阅主题”test/to”才能收到消息。

----------

+ **Q:** 为什么我对于需要执行的函数配置了运行实例，且对运行实例进行了资源限制，但是好像实际运行过程中并未起作用？

  **A:** 智能边缘规定对于运行实例的资源配置仅支持Linux系统，且要求系统内核版本高于3.10，以root权限启动，并在配置文件中启用（通过配置{control: {cgroup: true}}）才有效，其余情况下配置均无效。

----------

+ **Q:** 在云端管理平台配置了权限、规则和函数后下发至本地，reload导致本地原有remote配置失效，无法使用remote功能

  **A:** 该问题是由于目前云端可管理应用配置，包括权限配置（Principals）、路由规则配置（Subscriptions）、函数配置（Functions）及Remote配置（云端暂不支持Remote配置），下发后会覆盖本地的应用配置。非应用配置，比如Listen、Certificate、Cloud保持不变。

----------

+ **Q:** openedge无法启动，提示信息诸如Cloud.Address: zero value，Cloud.CA: zero value，Cloud.Username: zero value，Cloud.Password: zero value
![图片](http://agroup.baidu.com:8964/static/b4/d0cdb0a2f8f9dfee40ca0534d4508e2b4cebe1.png?filename=cloudq2.png)

  **A:** 配置文件中Cloud标签对应的字段缺失，必须配置的字段包括CA，Address，Username，Password。如果不使用Cloud功能，须移除整个Cloud配置项。

----------

+ **Q:** openedge启动后连接不上云管理，提示信息诸如Connect failed                              client=e45384e636ef4be59ee7a83bba5a514c component=mqtt_client error="MQTT client connects failed"
![图片](http://agroup.baidu.com:8964/static/4b/7bf9dafca4c14924c5084923682d4247bb4578.png?filename=cloudq3.png)

  **A:** 云管理连接失败，可能由网络不通造成，请检查外网是否可正常访问

----------

+ **Q:** openedge启动后连接不上云管理，提示信息诸如Connect failed                                client=e45384e636ef4be59ee7a83bba5a514c component=mqtt_client error="connection refused: bad user name or password"
![图片](http://agroup.baidu.com:8964/static/b9/442e5931f6f765fe782780c61a5079b17098a2.png?filename=cloudq4.png)

  **A:** 云管理连接失败，请检查配置文件中，cloud.username和cloud.password是否正确

----------

+ **Q:** openedge无法启动，提示信息诸如{Remotes[0].Protocol: regular expression mismatch}]
![图片](http://agroup.baidu.com:8964/static/6b/d8795a8b3ad05d7264055a669ad0cf8ac53609.png?filename=remoteq1.png)

  **A:** 配置文件中Remote配置有误，具体配置格式参见操作手册中的Remote章节

----------

+ **Q:** openedge配置了Remote功能，但启动后和远程服务的连接一直处于不断重连的状态。
![图片](http://agroup.baidu.com:8964/static/73/020b40562442eb0329370365a16dc7a787a7f1.png?filename=remoteq3.png)

  **A:** 可能原因是配置在Subscriptions中的Remote路由，没有对应的发布和订阅权限。请检查远程服务中是否开启了对应主题的发布和订阅权限。
