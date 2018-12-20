# Linux环境下编译OpenEdge

## 环境配置

具体环境配置请参考文档[环境配置](./OpenEdge-build-prepare.md)

## 源码下载

下载openedge源码

 ```shell
 mkdir -p $GOPATH/src/github.com/baidu/
 cd $GOPATH/src/github.com/baidu/
 git clone https://github.com/baidu/openedge.git
 ```

***注:*** openedge代码目录需要存放在 ```$GOPATH/src/github.com/baidu/``` 目录下

## 依赖拉取

```shell
cd $GOPATH/src/github.com/baidu/openedge
make depends
```

## 源码编译

```shell
cd $GOPATH/src/github.com/baidu/openedge
make
```

编译完成后会在根目录下生成如下四个可执行文件。

```shell
openedge
openedge-hub
openedge-function
openedge-remote-mqtt
```

## 程序安装

管理员身份安装openedge，并安装到默认路径：/usr/local。

```shell
cd $GOPATH/src/github.com/baidu/openedge
sudo make install # docker容器模式
sudo make native-install # native进程模式
```

非管理员身份安装openedge，并指定安装路径，比如docker容器模式安装到output/docker目录中，native进程模式安装到output/native目录中。

```shell
cd $GOPATH/src/github.com/baidu/openedge
make PREFIX=output/docker install # docker容器模式
make PREFIX=output/native native-install # native进程模式
```

## 程序运行

如果程序已经安装到默认路径：/usr/local。

```shell
openedge -w example/docker # docker容器模式
openedge # native进程模式
```

如果程序已经安装到了指定路径，比如docker容器模式安装到output/docker目录中，native进程模式安装到output/native目录中。

```shell
cd $GOPATH/src/github.com/baidu/openedge
output/docker/bin/openedge -w example/docker # docker容器模式
output/native/bin/openedge -w output/native # native进程模式
```

**提示**：

1. docker容器模式运行，可通过 ```docker stats``` 命令查看容器运行状态。
2. 如需使用自己的镜像，需要修改应用配置中的模块和函数的 entry，指定自己的镜像。
3. 如需自定义配置，请按照 [配置解读](../config/config.md) 中的内容进行相关设置。

### 程序卸载

如果是默认安装。

```shell
cd $GOPATH/src/github.com/baidu/openedge
make clean # 可用于清除编译生成的可执行文件
sudo make uninstall # docker容器模式
sudo make native-uninstall # native进程模式
```

如果是指定了安装路径，比如docker容器模式安装到output/docker目录中，native进程模式安装到output/native目录中。

```shell
cd $GOPATH/src/github.com/baidu/openedge
make clean # 可用于清除编译生成的可执行文件
make PREFIX=output/docker uninstall # docker容器模式
make PREFIX=output/native native-uninstall # native进程模式
```
