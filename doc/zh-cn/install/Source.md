# 从源码编译 OpenEdge

## Linux环境下编译

### 环境配置

具体环境配置请参考文档[环境配置](./Configurations.md#运行环境配置)。

### 源码下载

下载openedge源码

 ```shell
 mkdir -p $GOPATH/src/github.com/baidu/
 cd $GOPATH/src/github.com/baidu/
 git clone https://github.com/baidu/openedge.git
 ```

***注:*** openedge代码目录需要存放在 ```$GOPATH/src/github.com/baidu/``` 目录下

### 源码编译

```shell
cd $GOPATH/src/github.com/baidu/openedge
make # 编译主程序和模块的可执行程序
make images # 在本地生成模块镜像
```

编译完成后会在根目录下生成如下四个可执行文件。

```shell
openedge
openedge-hub
openedge-function
openedge-remote-mqtt
```

### 程序安装

安装到默认路径：/usr/local。

```shell
cd $GOPATH/src/github.com/baidu/openedge
make install
```

指定安装路径，比如安装到output目录中。

```shell
cd $GOPATH/src/github.com/baidu/openedge
make PREFIX=output install
```

### 程序运行

如果程序已经安装到默认路径：/usr/local。

```shell
openedge -w example/docker # docker容器模式
openedge # native进程模式
```

如果程序已经安装到了指定路径，比如安装到output目录中。

```shell
cd $GOPATH/src/github.com/baidu/openedge
output/bin/openedge -w example/docker # docker容器模式
output/bin/openedge # native进程模式
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
make uninstall
```

如果是指定了安装路径，比如安装到output目录中。

```shell
cd $GOPATH/src/github.com/baidu/openedge
make clean # 可用于清除编译生成的可执行文件
make PREFIX=output uninstall
```




