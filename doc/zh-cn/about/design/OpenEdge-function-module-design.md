# 函数计算模块（openedge_function）

函数计算提供基于MQTT消息机制，弹性、高可用、扩展性好、响应快的的计算能力，并且兼容[百度云-函数计算CFC](https://cloud.baidu.com/product/cfc.html)。函数通过一个或多个具体的实例执行，每个实例都是一个独立的进程，现采用grpc server运行函数实例。所有函数实例由实例池（pool）负责管理生命周期，支持自动扩容和缩容。结构图如下：

![函数计算模块结构图](../../images/about/function.png)

**注意**: 如果函数执行错误，函数计算会返回如下格式的消息，供后续处理。其中packet是函数输入的消息（被处理的消息），不是函数返回的消息

```python
{
    "errorMessage": "rpc error: code = Unknown desc = Exception calling application",
    "errorType": "*errors.Err",
    "packet": {
        "Message": {
            "Topic": "t",
            "Payload": "eyJpZCI6MSwiZGV2aWNlIjoiMTExIn0=",
            "QOS": 0,
            "Retain": false
        },
        "Dup": false,
        "ID": 0
    }
}
```

## 函数计算python2.7 runtime模块（openedge_function_runtime_python2.7）

Python函数与[百度云-函数计算CFC](https://cloud.baidu.com/product/cfc.html)类似，用户通过编写的自己的函数来处理消息，可进行消息的过滤、转换和转发等，使用非常灵活。

Python函数的输入输出可以是JSON格式也可以是二进制形式。消息payload在作为参数传给函数前会尝试一次JSON解码（json.loads(payload)），如果成功则传入字典（dict）类型，失败则传入原二进制数据。

Python函数支持读取环境变量，比如os.environ['PATH']。

Python函数支持读取上下文，比如context['functionName']。

Python函数实现举例：

```python
#!/usr/bin/env python
#-*- coding:utf-8 -*-
"""
module to say hi
"""

def handler(event, context):
    """
    function handler
    """
    event['functionName'] = context['functionName']
    event['functionInvokeID'] = context['functionInvokeID']
    event['functionInstanceID'] = context['functionInstanceID']
    event['messageQOS'] = context['messageQOS']
    event['messageTopic'] = context['messageTopic']
    event['sayhi'] = '你好，世界！'
    return event
```

> **提示**：Native进程模式下，若要运行本代码库提供的[sayhi.py](https://github.com/baidu/openedge/blob/master/example/native/app/func-nyeosbbch/sayhi.py)，需要自行安装python2.7，且需要基于python2.7安装protobuf3、grpcio、pyyaml(采用pip安装即可，`pip install grpcio protobuf pyyaml`)。

> 此外，对于Native进程模式python脚本运行环境的构建，推荐通过virtualenv构建虚拟环境，并在此基础上安装相关依赖，相关步骤如下：
>
> ```python
>   # install virtualenv via pip
>   [sudo] pip install virtualenv
>   # test your installation
>   virtualenv --version
>   # build workdir run environment
>   cd /path/to/native/workdir
>   virtualenv native/workdir
>   # install requirements
>   source bin/activate # activate virtualenv
>   pip install grpcio protobuf pyyaml # install grpc, protobuf3, yaml via pip
>   # test native mode
>   bin/openedge # run openedge under native mode
>   deactivate # deactivate virtualenv
> ```

## 函数计算node8.5 runtime模块（openedge_function_runtime_node85）

Javascript函数与[百度云-函数计算CFC](https://cloud.baidu.com/product/cfc.html)类似，用户通过编写的自己的函数来处理消息，可进行消息的过滤、转换和转发等，使用非常灵活。

Javascript函数的输入输出可以是JSON格式也可以是二进制形式。消息payload在作为参数传给函数前会尝试一次JSON解码，如果成功则传入Json类型，失败则传入原二进制数据。

Javascript函数实现举例：

```JavaScript
#!/usr/bin/env node
exports.handler = (event, context, callback) => {
    event.name = 'openedge';
    event.say = 'hi';
    callback(null, event);
};
```

> **提示**：Native进程模式下，若要运行本代码库提供的[sayhi.js](https://github.com/baidu/openedge/tree/master/example/native/app/func-qwertyuia/sayhi.js)，需要自行安装node8.5和相关依赖[package.json文件中列出]。

> 首先安装node8.5，后安装依赖，相关步骤如下：
>
> ```shell
>   # build workdir run environment
>   cd /path/to/native/workdir/bin
>   # install dependencies via package
>   npm install
>   # test native mode
>   openedge # run openedge under native mode
> ```