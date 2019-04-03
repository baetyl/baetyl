# 自定义函数运行时

- [协议约定](#协议约定)
- [启动约定](#启动约定)

函数运行时是函数执行的载体，通过动态加载函数代码来执行函数，跟函数实现的语言强相关。比如 Python 代码需要使用 Python 运行时来调用。这里就涉及到多种语言的问题，为了统一调用接口和协议，我们最终选择了 GRPC，借助其强大的跨语言 IDL 和高性能 RPC 通讯能力，打造可灵活扩展的函数计算框架。

在函数计算服务中，`openedge-function-manager` 负责函数实例的管理和调用。函数实例由函数运行时服务提供，函数运行时服务只需满足下面介绍的约定。

## 协议约定

开发者可直接使用 sdk/openedge-go 中的 `function.proto` 生成各自编程语言的消息和服务实现，具体定义如下。GRPC 使用方法可参考 [GRPC 官网的文档](https://grpc.io/docs/quickstart/go.html)。

```proto
syntax = "proto3";

package openedge;

// 函数 Server 定义
service Function {
  rpc Call(FunctionMessage) returns (FunctionMessage) {}
  // rpc Talk(stream Message) returns (stream Message) {}
}

// 函数调用和返回的消息体
message FunctionMessage {
  uint64 ID                 = 1;
  uint32 QOS                = 2;
  string Topic              = 3;
  bytes  Payload            = 4;

  string  FunctionName      = 11;
  string  FunctionInvokeID  = 12;
}
```

**注意**：Docker 容器模式下，函数实例的资源限制不要低于 `50M` 内存，`20` 个线程。

## 配置约定

函数运行时模块并不强制约定配置，但是为了统一配置方式，推荐如下配置项

- name：函数名称
- handler：函数处理接口
- codedir：函数代码所在路径，如果有的话。

下面是一个 Python 函数运行时服务的配置举例：

```yaml
functions:
  - name: 'sayhi'
    handler: 'sayhi.handler'
    codedir: 'var/db/openedge/function-sayhi'
```

## 启动约定

函数运行时服务同其他服务一样，唯一的区别是实例是由其他服务动态启动的。比如为了避免监听端口冲突，可以动态指定端口。函数运行时模块可以从环境变量中读取 `OPENEDGE_SERVICE_ADDRESS` 作为 GRPC Server 监听的地址。另外，动态启动的函数实例没有权限调用主程序的API。最后，模块监听 `SIGTERM` 信号来实现优雅退出。完整的实现可参考 Python2.7 运行时模块（`openedge-function-python27`）。
