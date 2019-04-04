# Customize Runtime Module

- [Protocol Convention](#protocol-convention)
- [Configuration Convention](#configuration-convention)
- [Start/Stop Convention](#startstop-convention)

The function runtime is the carrier of the function execution. The function is executed by dynamically loading the function code, which is strongly related to the language of the function implementation. For example, Python code needs to be called using the Python runtime. This is a multi-language issue. In order to unify the interface and protocol, we finally chose GRPC to create a flexible functional computing framework with its powerful cross-language IDL and high-performance RPC communication capabilities.

In the function compute service (FaaS), `openedge-function-manager` is responsible for the management and invocation of function instances. The function instance is provided by the function runtime service, and the function runtime service only needs to meet the conventions described below.

## Protocol Convention

Developers can use the `function.proto` in `sdk/openedge-go` to generate messages and service implementations for their respective programming languages, as defined below. For the usage of GRPC, refer to [GRPC Official Documents](https://grpc.io/docs/quickstart/go.html).

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

_**NOTE**: In docker container mode, the resource limit of the function instance should not be lower than `50M` memory and 20 threads._

## Configuration Convention

The function runtime module does not enforce the configuration. However, for the unified configuration mode, the following configuration items are recommended.

- name: function name
- handler: function processing interface
- codedir: The path to the function code, if any.

The following is a configuration example of a Python function runtime service:

```yaml
functions:
  - name: 'sayhi'
    handler: 'sayhi.handler'
    codedir: 'var/db/openedge/function-sayhi'
```

## Start/Stop Convention

The function runtime service is the same as other services, the only difference is that the instance is dynamically started by other services. For example, to avoid listening port conflicts, you can specify the port dynamically. The function runtime module can read `OPENEDGE_SERVICE_ADDRESS` from the environment variable as the address that the GRPC Server listens on. In addition, dynamically launched function instances do not have permission to call the main program's API. Finally, the module listens for the `SIGTERM` signal to gracefully exit. A complete implementation can be found in the Python 2.7 runtime module (`openedge-function-python27`).