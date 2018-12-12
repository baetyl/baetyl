# Linux环境下编译OpenEdge

## 环境配置

具体环境配置请参考文档[环境配置](./OpenEdge-build-prepare.md)

## 源码拉取

执行命令 ```git clone https://github.com/baidu/openedge.git``` 拉取最新的Openedge 代码。

***注:*** openedge代码目录需要存放在 ```$GOPATH/src/github.com/baidu/``` 目录下

## 依赖拉取

在环境配置完成后，进入openedge源码根目录下，执行 ```make depends``` 命令拉取所需依赖。完成后进入下一步。

## 源码编译

### 编译并发布镜像

1. 登录镜像仓库```docker login 镜像仓库```

2. 进入openedge源码根目录，执行命令 ```sh image.sh -v 版本号 -r 镜像仓库```，等待编译完成即可。

### 编译openedge主程序

进入openedge源码根目录，执行命令 ```make``` 即可开始编译，编译完成后会在根目录下生成 ```openedge```, ```openedge-hub```, ```openedge-function```, ```openedge-remote-mqtt``` 四个可执行文件。

### 安装

#### docker 模式安装

如您计划以 docker 方式启动ooenedge，执行命令 ```sudo make install```，输入管理员密码后即可完成安装

#### native 模式安装

如您计划以 native 方式启动openedge，执行命令 ```sudo make native-install```，输入管理员密码后即可完成安装

### 启动

按照对应模式完成安装后，即可进入启动阶段。

#### docker 模式启动

如您计划以 docker 方式启动，在按照[docker 模式安装](#docker-模式安装)后，执行命令 ```openedge -w example/docker```，即可按照所提供示例配置以docker模式启动openedge。可通过 ```docker stats``` 命令查看容器运行状态。

#### native 模式启动

如您计划以 native 方式启动，首先需要按照[native 模式安装](#native-模式安装)步骤完成基本的安装操作，完成后执行命令 ```openedge``` 即可成功启动。

### 测试

按照[启动](#启动)下所描述的内容完成启动操作后，可以进行简单功能的测试。

在Openedge根目录下执行命令```make test```, 会生成 ```benchmark```, ```consistency```, ```pubsub``` 三个可执行文件。

我们这里通过pubsub命令来做一些简单的测试:

1. 在终端下进入Openedge根目录，通过命令 ```./pubsub sub -a tcp://127.0.0.1:1883 -u test -p hahaha -t t``` 来启动pubsub程序进行订阅，其中用户名和密码可以在hub配置中进行修改，具体请参考[配置解读](../config/config.md)。
2. 新建一个终端，同样进入Openedge根目录通过命令 ```./pubsub pub -a tcp://127.0.0.1:1883 -u test -p hahaha -t t``` 来启动pubsub程序并进行pub操作，程序启动后，我们可以继续输入想要推送的内容，如果订阅使用的终端收到消息，此次测试通过。

> 其他测试可通过 ```-h```参数查看相关内容并进行相关测试。

### 卸载

切换到openedge根目录:

+ 执行 ```make clean```命令可以清除根目录下因编译过程生成的可执行文件
+ 如需彻底卸载openedge相关程序，执行命令 ```make uninstall``` (docker 模式卸载) 或命令 ```make native-uninstall``` (native 模式卸载) 即可完成卸载。

**提示**：

1. 如需使用自己的镜像，需要修改应用配置中的模块和函数的 entry，指定自己的镜像。
2. 如需自定义配置，请按照 [配置解读](../config/config.md) 中的内容进行相关设置。
