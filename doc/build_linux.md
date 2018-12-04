# linux环境下编译openedge

- [环境配置](#环境配置)
- [依赖拉取](#依赖拉取)
- [源码编译](#源码编译)
  - [编译并发布镜像](#编译并发布镜像)
  - [编译openedge主程序](#编译openedge主程序)
  - [native模式启动openedge](#native模式启动openedge)
  - [docker模式启动openedge](#docker模式启动openedge)

## 环境配置

具体环境配置请参考文档[环境配置](./build_prepare.md)

## 依赖拉取

在环境配置完成后，进入openedge源码根目录下，执行 **godep restore** 拉取所需依赖。完成后进入下一步。

## 源码编译

### 编译并发布镜像

1. 登录镜像仓库```docker login 镜像仓库```

2. 进入openedge源码根目录，执行命令 ```sh image.sh -v 版本号 -r 镜像仓库```，等待编译完成即可。

### 编译openedge主程序

进入openedge源码根目录，执行命令 ```sh build.sh``` 等待编译完成即可。

***注*** ：Linux环境需要注意sh所链接命令是否为bash，使用命令```readlink -f $(which sh)``` 即可查看，build.sh需要使用bash来运行。

### native模式启动openedge

进入openedge源码根目录，执行```./output/native/bin/openedge -w output/native```

### docker模式启动openedge

进入openedge源码根目录，执行```./output/docker/bin/openedge -w output/docker```

**提示**：如需使用自己的镜像，需要修改[应用配置](../example/docker/app/app.yml)中的模块和函数的 entry，指定自己的镜像。
