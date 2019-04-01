# 自定义函数运行时

- [自定义函数运行时](#自定义函数运行时)
  - [协议约定](#协议约定)
  - [启动约定](#启动约定)

函数运行时是函数执行的载体，通过加载函数代码来调用函数，跟函数实现的语言强相关。比如 Python 代码需要使用 Python 运行时来调用。这里就涉及到多种语言的问题，为了统一调用接口和协议，我们最终选择了 GRPC，借助其强大的跨语言 IDL 和高性能 RPC 通讯能力，打造可灵活扩展的函数计算框架。

在函数计算服务中，openedge-funtcion-manager负责函数实例的管理和调用。函数实例由函数运行时服务提供，函数运行时服务只需满足下面介绍的约定。

## 协议约定

开发者可直接使用sdk/openedge-go中的function.proto生成各自编程语言的消息和服务实现，具体定义如下。GRPC使用方法可参考 [GRPC 官网的文档](https://grpc.io/docs/quickstart/go.html)。

```proto
syntax = "proto3";

package openedge;

// The function server definition.
service Function {
  rpc Call(FunctionMessage) returns (FunctionMessage) {}
  // rpc Talk(stream Message) returns (stream Message) {}
}

// FunctionMessage function message
message FunctionMessage {
  uint64 ID                 = 1;
  uint32 QOS                = 2;
  string Topic              = 3;
  bytes  Payload            = 4;

  string  FunctionName      = 11;
  string  FunctionInvokeID  = 12;
}
```

**注意**：Docker 容器模式下，函数实例的资源限制不要低于 `50M` 内存，20 个线程。

## 启动约定

函数运行时服务同其他服务一样，唯一的区别是示例是被其他服务动态启动的。比如为了避免监听端口的冲突，可以动态指定端口。函数运行时模块约定从环境变量中读取`OPENEDGE_SERVICE_ADDRESS`作为GPRC Server监听的地址。另外动态启动的函数实例是没有权限调用主程序的API的。最后注意监听 `SIGTERM` 信号来优雅退出。完整的实现过程可参考 Python2.7 运行时（openedge-function-python27）。
