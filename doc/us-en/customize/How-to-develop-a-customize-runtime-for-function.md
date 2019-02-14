# How to develop a customize runtime for function

The function runtime is the carrier of the function execution, which is strongly related to the language of the function script. For example, python script needs to be called using the Python2.7 runtime. In order to solve the multi-language issues and unify the interface and protocol, we finally chose GRPC, and with its powerful cross-language IDL and high-performance RPC communication capabilities, we can create a flexible function computing framework.

## Protocol Convention

[GRPC message and service definition](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/module/function/runtime/openedge_function_runtime.proto) is as follows, developers can directly use .pb file to generate messages and service implementations in their respective programming languages. For GRPC usage, refer to [```make pb``` command](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/Makefile), or [GRPC official documentation](https://grpc.io/docs/quickstart/go.html)。

```proto
syntax = "proto3";

package runtime;

// The runtime definition.
service Runtime {
  // Handle handles request
  rpc Handle (Message) returns (Message) {} // message handling interface
}

// The request message.
message Message {
  uint32 QOS                = 1; // QOS of MQTT message
  string Topic              = 2; // Topic of MQTT message
  bytes  Payload            = 3; // Payload of MQTT message

  string FunctionName       = 11; // Function name invoked
  string FunctionInvokeID   = 12; // Function invoke id
}
```

_**NOTE**: In docker container mode, the resource limit of the function instance should not be lower than 50M memory and 20 threads._

## Configuration Convention

[All definitions of function configuration](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/module/config/function.go) are as follows. For the custom function runtime, just pay attention to the definition of Function.Handler and Function.CodeDir. Other configurations are only used by the function framework.

```golang

// Runtime runtime config
type Runtime struct {
    // module configuration
    Module   `yaml:",inline" json:",inline"`
    // server configuration
    Server   RuntimeServer `yaml:"server" json:"server"`
    // function configuration
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
    // function ID used to map the directory of the function runtime work directory
    ID      string `yaml:"id" json:"id" validate:"regexp=^[a-zA-Z0-9_-]{1\\,140}$"`
    // function name used to invoke
    Name    string `yaml:"name" json:"name" validate:"regexp=^[a-zA-Z0-9_-]{1\\,140}$"`
    // funciton invoke entry
    Handler string `yaml:"handler" json:"handler" validate:"nonzero"`
    // function code directory
	CodeDir string `yaml:"codedir" json:"codedir"`
    // instance configuration
    Instance Instance          `yaml:"instance" json:"instance"`
    // function runtime, it is docker image in docker container mode and is executable program path in native process mode
    Entry    string            `yaml:"entry" json:"entry"`
    // function runtime env
	Env      map[string]string `yaml:"env" json:"env"`
}

// Instance instance config for function runtime module
type Instance struct {
    // minimum number of instances
    Min       int           `yaml:"min" json:"min" default:"0" validate:"min=0, max=100"`
    // maximum number of instances
    Max       int           `yaml:"max" json:"max" default:"1" validate:"min=1, max=100"`
    // maximum idle time
    IdleTime  time.Duration `yaml:"idletime" json:"idletime" default:"10m"`
    // invoke timeout
    Timeout   time.Duration `yaml:"timeout" json:"timeout" default:"5m"`
    // resources configuration
    Resources Resources     `yaml:",inline"  json:",inline"`
    // maximum length of message
	  Message   struct {
		    Length Length `yaml:"length" json:"length" default:"{\"max\":4194304}"`
	  } `yaml:"message" json:"message"`
}
```

## Start/Stop Convention

The function runtime is actually a module, but it is special. It is a module that is started by the function module during running and is classified as a temporary module. The start/stop convention is basically the same as that of the resident module, except that the temporary module does not have a fixed path and does not support loading the configuration from the configuration file. Therefore, it is agreed to pass configuration in through process arguments in JSON format, such as -c "{\"name\":\"sayhi\", ...}".

After the function runtime starts and loads the configuration, first initialize the GRPC server according to the server configuration, then load the function code entry according to the configuration of the Function, wait for the function module to call, and finally monitor the SIGTERM signal to gracefully exit. The complete implementation can refer to [python27 runtime](https://github.com/baidu/openedge/tree/5010a0d8a4fc56241d5febbc03fdf1b3ec28905e/openedge-function-runtime-python27/openedge_function_runtime_python27.py)。
