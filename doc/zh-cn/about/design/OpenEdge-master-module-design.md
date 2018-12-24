# 主程序(master)

[主程序](https://github.com/baidu/openedge/blob/master/master/master.go)负责所有模块的管理、云同步等，由模块引擎、云代理和API构成。

## 模块引擎(engine)

[模块引擎](https://github.com/baidu/openedge/blob/master/engine/engine.go)负责模块的启动、停止、重启、监听和守护，目前支持Docker容器模式和Native进程模式。

模块引擎从工作目录的[var/db/openedge/module/module.yml](https://github.com/baidu/openedge/blob/master/example/docker/var/db/openedge/module/module.yml)配置中加载模块列表，并以列表的顺序逐个启动模块。模块引擎会为每个模块启动一个守护协程对模块状态进行监听，如果模块异常退出，会根据模块的[Restart Policy](https://github.com/baidu/openedge/blob/master/module/config/policy.go)配置项执行重启或退出。主程序关闭后模块引擎会按照列表的逆序逐个关闭模块。

_**提示**：工作目录可在OpenEdge启动时通过-w指定，默认为OpenEdge的可执行文件所在目录的上一级目录。_

Docker容器模式下，模块通过docker client启动entry指定的模块镜像，并自动加入自定义网络（openedge）中，由于docker自带了DNS server，模块间可通过模块名进行互相通讯。另外模块可通过expose配置项来向外暴露端口；通过resources配置项限制模块可使用的资源，目前支持CPU、内存和进程数的限制。[参考配置](https://github.com/baidu/openedge/blob/master/example/docker/var/db/openedge/module/module.yml)

Docker容器模式下，模块在容器中的工作目录是根目录：/，配置路径是：/etc/openedge/module.yml，资源读取目录是：/var/db/openedge/module/<模块名>，持久化数据输出目录是：/var/db/openedge/volume/<模块名>，日志输出目录是：/var/log/openedge/<模块名>。具体的映射如下：

> - \<openedge_host_work_dir>/var/db/openedge/module/<module_name>/module.yml:/etc/openedge/module.yml
> - \<openedge_host_work_dir>/var/db/openedge/module/<module_name>:/var/db/openedge/module/<module_name>
> - \<openedge_host_work_dir>/var/db/openedge/volume/<module_name>:/var/db/openedge/volume/<module_name>
> - \<openedge_host_work_dir>/var/log/openedge/<module_name>:/var/log/openedge/<module_name>

_**提示**：应用模块在云端被修改后mark值会发生变化，和git的修订版本类似。_

Native进程模式下，通过syscall启动entry指定模块可执行文件，模块间可通过localhost（127.0.0.1）进行互相通讯，不支持资源隔离和资源限制。

## 云代理(agent)

云代理负责和云端管理套件通讯，走MQTT和HTTPS通道，MQTT强制SSL/TLS证书双向认证，HTTPS强制SSL/TLS证书单向认证。

OpenEdge启动和热加载（reload）完成后会通过云代理上报一次设备信息，目前[上报的内容](https://github.com/baidu/openedge/blob/master/agent/report.go)如下：

> - go_version：OpenEdge主程序的Golang版本
> - bin_version：OpenEdge主程序的版本
> - conf_version：OpenEdge主程序加载应用的版本
> - reload_error：OpenEdge主程序加载应用的报错信息，如果有的话
> - os：设备的系统，比如：linux、windows、darwin
> - bit：设备CPU位数，比如：32、64
> - arch：设备CPU架构，比如：amd64
> - gpu\<n\>：设备的第n个GPU的型号
> - gpu\<n\>_mem_total：设备的第n个GPU的显存总容量
> - gpu\<n\>_mem_free：设备的第n个GPU的显存剩余容量
> - mem_total：设备的内存总容量
> - mem_free：设备的内存剩余容量
> - swap_total：设备的交换空间总容量
> - swap_free：设备的交换空间剩余容量

云代理接收到云端管理套件的应用下发指令后，OpenEdge开始执行[热加载](https://github.com/baidu/openedge/blob/master/master/master.go)，流程如下图：

![热加载流程](../../images/about/reload.png)

_**注意**：目前热加载采用全量更新的方式，既先停止所有老应用模块再启动所有新应用模块，因此模块提供的服务会中断。另外解压新应用包会直接覆盖老应用的文件，多余的老应用文件不会被清理，后续会考虑加入清理逻辑。_

_**提示**：如果设备无法连接外网或者脱离云端管理，可移除cloud配置项，离线运行。_


## API(api)

OpenEdge主程序会暴露一组HTTP API，目前支持获取空闲端口，模块的启动、停止和重启。API server在linux系统下默认采用unix domain socket，工作目录下的var/run/openedge.sock；其他环境采用TCP，默认监听tcp://127.0.0.1:50050。开发者可以通过[api](https://github.com/baidu/openedge/blob/master/config/config.go)配置项修改监听的地址。为了方便管理，我们对模块做了一个划分，从var/db/openedge/module/module.yml中加载的模块称为常驻模块，通过API启动的模块称为临时模块，临时模块遵循谁启动谁负责停止的原则。OpenEdge退出时，会先逆序停止所有常驻模块，常驻模块停止过程中也会调用API来停止其启动的模块，最后如果还有遗漏的临时模块，会随机全部停止。访问API需要提供账号和密码，设置如下两个请求头：

> - x-iot-edge-username：账号名称，即常驻模块名
> - x-iot-edge-password：账号密码，即常驻模块的token，使用```module.GetEnv(module.EnvOpenEdgeModuleToken)```从环境变量中获取，OpenEdge主程序启动时会为每个常驻模块生成临时的token。

官方提供的函数计算模块会调用这个API来启停函数实例（runtime）模块，函数计算模块停止时需要负责将其启动的临时模块全部停掉。

## 环境变量(env)

OpenEdge目前会给模块设置如下几个系统环境变量：

> - OPENEDGE_HOST_OS：OpenEdge所在设备（宿主机）的系统类型
> - OPENEDGE_MASTER_API：OpenEdge主程序的API地址
> - OPENEDGE_MODULE_MODE：OpenEdge主程序的运行模式
> - OPENEDGE_MODULE_TOKEN：OpenEdge主程序分配给常驻模块的临时token，可作为常驻模块访问主程序API的密码

官方提供的函数计算模块就是通过读取OPENEDGE_MASTER_API来连接OpenEdge主程序的，比如linux系统下OPENEDGE_MASTER_API默认是unix://var/run/openedge.sock；其他系统的docker容器中OPENEDGE_MASTER_API默认是tcp://host.docker.internal:50050；其他系统的native模式下OPENEDGE_MASTER_API默认是tcp://127.0.0.1:50050。

_**注意**：应用中配置的环境变量如果和上述系统环境变量相同会被覆盖。_
