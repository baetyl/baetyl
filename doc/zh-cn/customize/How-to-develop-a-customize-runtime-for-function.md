# 自定义函数运行时

- [自定义函数运行时](#自定义函数运行时)
  - [协议约定](#协议约定)
  - [配置约定](#配置约定)
  - [启动约定](#启动约定)

函数运行时是函数执行的载体，通过加载函数代码来调用函数，跟函数实现的语言强相关。比如python代码需要使用python runtime来调用。这里就涉及到多种语言的问题，为了统一调用接口和协议，我们最终选择了GRPC，借助其强大的跨语言IDL和高性能RPC通讯能力，打造可灵活扩展的函数计算框架。

## 协议约定

[GRPC的消息和服务定义](https://github.com/baidu/openedge/blob/master/module/function/runtime/openedge_function_runtime.proto)如下，开发者可直接用于生成各自编程语言的消息和服务实现，GRPC使用方法可参考[```make pb```命令](https://github.com/baidu/openedge/blob/master/Makefile)，也可参考[GRPC官网的文档](https://grpc.io/docs/quickstart/go.html)。

```proto
syntax = "proto3";

package runtime;

// The runtime definition.
service Runtime {
  // Handle handles request
  rpc Handle (Message) returns (Message) {} // 消息处理接口
}

// The request message.
message Message {
  uint32 QOS                = 1; // MQTT消息的QOS
  string Topic              = 2; // MQTT消息的主题
  bytes  Payload            = 3; // MQTT消息的内容

  string FunctionName       = 11; // 被调用的函数名
  string FunctionInvokeID   = 12; // 函数调用ID
}
```

## 配置约定

[函数配置的所有定义](https://github.com/baidu/openedge/blob/master/module/config/function.go)如下。对于自定义函数runtime，只需关注Function.Handler和Function.CodeDir的定义和使用，其他配置都是函数计算框架使用的配置。

```golang

// Runtime runtime config
type Runtime struct {
    // 模块哦诶只
    Module   `yaml:",inline" json:",inline"`
    // 服务配置
    Server   RuntimeServer `yaml:"server" json:"server"`
    //
	Function Function      `yaml:"function" json:"function"`
}

// RuntimeServer function runtime server config
type RuntimeServer struct {
	Address string        `yaml:"address" json:"address" validate:"nonzero"`
	Timeout time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
	Message struct {
		Length Length `yaml:"length" json:"length" default:"{\"max\":4194304}"`
	} `yaml:"message" json:"message"`
}

// Function function config
type Function struct {
    // 函数ID，目前用于映射函数代码所在的目录
    ID      string `yaml:"id" json:"id" validate:"regexp=^[a-zA-Z0-9_-]{1\\,140}$"`
    // 函数名，函数调用使用的名称
    Name    string `yaml:"name" json:"name" validate:"regexp=^[a-zA-Z0-9_-]{1\\,140}$"`
    // 函数执行的入口
    Handler string `yaml:"handler" json:"handler" validate:"nonzero"`
    // 函数代码所在目录
	CodeDir string `yaml:"codedir" json:"codedir"`
    // 实例配置
    Instance Instance          `yaml:"instance" json:"instance"`
    // 函数runtime，docker容器模式下是函数runtime的镜像，native进程模式下是函数runtime的可执行程序路径
    Entry    string            `yaml:"entry" json:"entry"`
    // 函数runtime启动进程时传入的环境变量
	Env      map[string]string `yaml:"env" json:"env"`
}

// Instance instance config for function runtime module
type Instance struct {
    // 最少实例数
    Min       int           `yaml:"min" json:"min" default:"0" validate:"min=0, max=100"`
    // 最多实例数
    Max       int           `yaml:"max" json:"max" default:"1" validate:"min=1, max=100"`
    // 实例最大空闲时间
    IdleTime  time.Duration `yaml:"idletime" json:"idletime" default:"10m"`
    // 实例调用超时时间
    Timeout   time.Duration `yaml:"timeout" json:"timeout" default:"5m"`
    // 实例的资源配置
    Resources Resources     `yaml:",inline"  json:",inline"`
    // 消息最大长度限制
	Message   struct {
		Length Length `yaml:"length" json:"length" default:"{\"max\":4194304}"`
	} `yaml:"message" json:"message"`
}
```

## 启动约定

函数runtime实际上也是一种模块，只是比较特殊，是被函数计算模块在运行过程中按需启动的模块，归为临时模块。其启动方式和普通模块启动的方式基本一致，只不过临时模块没有固定路径，不支持从配置文件中加载配置。因此约定通过传参的方式传入配置信息，使用JSON格式，比如-c "{\"name\":\"sayhi\", ...}"。

函数runtime启动并加载配置后，首先根据Server配置初始化GRPC server，然后根据Function的配置加载函数代码入口，等待函数计算模块来调用，最后注意监听SIGTERM信号来优雅退出。完整的实现过程可参考[python27 runtime](https://github.com/baidu/openedge/blob/master/openedge-function-runtime-python27/openedge_function_runtime_python27.py)。
