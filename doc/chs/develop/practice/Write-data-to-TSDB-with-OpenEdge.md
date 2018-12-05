# 测试前准备

**声明**：本文测试所用设备系统为MacOS，模拟MQTT client行为的客户端为[MQTTBOX](http://workswithweb.com/html/mqttbox/downloads.html)。

## 基本步骤流程

通过OpenEdge将数据写入TSDB具体依OpenEdge的函数计算服务实现，具体包括datapoint构造、身份信息签名及数据写入3个部分。

### datapoint构造

```python
def build_datapoints(event):
    """
    function to build datapoints by event
    datapoint for example: {
        "datetime": "2018-08-10 18:15:05",
        "temperature": 32,
        "unit": "℃"
    }
    """
    datapoints = dict()
    datapoint = dict()
    datapoint['metric'] = 'temperature'
    datapoint['tags'] = {'unit': event['unit']}
    datapoint['value'] = event['temperature']
    timestamp = time.mktime(time.strptime(event['datetime'],
                            '%Y-%m-%d %H:%M:%S'))
    datapoint['timestamp'] = str(timestamp).split('.')[0]
    datapoints['datapoints'] = [datapoint]
    return datapoints
```

通过上述代码即可成功构造datapoint数据（dict字典类型），其中metric、tags字段为**必选**字段，即构造的datapoint数据中必须包含metric和tags（TSDB要求，具体可查看[TSDB API](https://cloud.baidu.com/doc/TSDB/API/39.5C.E6.95.B0.E6.8D.AEAPI.E6.8E.A5.E5.8F.A3.E8.AF.B4.E6.98.8E.html)细节）。

### 身份信息签名

```python
#!/usr/bin/env python
# -*- coding: utf-8 -*-
	
import calendar
import datetime
import json
import requests
import time
	
from sign import sign
	
# set the transport protocol
TRANS_PROTOCOL = 'http://'
	
# set http method
HTTP_METHOD = 'POST'
	
# set base url and path
base_url = 'your_db.tsdb.iot.xx.baidubce.com'
path = '/v1/datapoint' # write your_data to datapoint of your_db on TSDB
	
# save the information of Access Key ID and Secret Access Key
ak = 'your_ak_info'
sk = 'your_sk_info'
credentials = sign.BceCredentials(ak, sk)
	
# set a http header except field 'Authorization'
headers = {'Host': base_url, 'Content-Type': 'application/json;charset=utf-8'}
	           
# we don't have params in our url,so set it to None
	
# set header fields should be signed headers_to_sign = {"host"}
	
# invoke sign method to get a signed string
sign_str = sign.sign(credentials, HTTP_METHOD, path, headers, params, headers_to_sign=headers_to_sign)
```

不难看出，通过上述代码即可完成身份信息的签名，需要注意的是，对身份信息签名时需要用户在百度云注册账户，并创建属于自己的AK/SK信息，具体可参考[如何获取AK/SK](https://cloud.baidu.com/doc/Reference/GetAKSK.html#.E5.A6.82.E4.BD.95.E8.8E.B7.E5.8F.96AK.20.2F.20SK)。

### 数据写入

```python
def access_db(http_method, url, data=None):
    """
    function to access TSDB by RESTful API（only have GET,POST,PUT now)
    """

    # invoke sign method to get a signed string
    sign_str = sign.sign(credentials, HTTP_METHOD, path, headers, params,
                         headers_to_sign=headers_to_sign)

    # add field 'Authorization' to complete the whole http header
    final_headers = dict(headers.items() + {'Authorization': sign_str}.items())

    try:
        if (http_method == 'POST') and (data is not None):
            rsp = requests.post(url, headers=final_headers, data=json.dumps(data))
        elif http_method == 'GET':
            rsp = requests.get(url, headers=final_headers)
        elif (http_method == 'PUT') and (data is not None):
            rsp = requests.put(url, headers=final_headers, data=json.dumps(data))
        else:
            rsp = 'Bad http method or data is empty'
    except StandardError:
        raise

    return rsp

def handler(event, context):
    """
    function handler
    """

    datapoints = build_datapoints(event)
    try:
        rsp = access_db(HTTP_METHOD, TRANS_PROTOCOL + base_url + path,
                        datapoints)
    except StandardError:
        raise

    # check http response status code to confirm if we write data successfully
    if str(rsp.status_code) == '204':
        pass
    else:
        if isinstance(rsp, str):
            raise TypeError('Response must be a string')
        else:
            raise BaseException('Get error: ' + str(rsp.status_code))
```

上述函数方法中，access\_db()方法将身份签名信息联合构造的datapoint数据信息一起通过POST方法写入TSDB（GET、PUT方法具体用法请参考[TSDB API](https://cloud.baidu.com/doc/TSDB/API.html#.E6.9F.A5.E8.AF.A2data.20point)说明）;handler()方法是整体程序的入口，用于调用access_db()方法将数据写入TSDB，并通过写入返回的状态信息判断数据是否写入成功。

## 测试与验证

```yaml
OpenEdge Hub模块配置
name: openedge_hub
mark: modu-nje2uoa9s
listen:
  - tcp://:1883
  - ssl://:1884
  - ws://:8080/mqtt
  - wss://:8884/mqtt
certificate:
  ca: 'app/cert-4j5vze02r/ca.pem'
  cert: 'app/cert-4j5vze02r/server.pem'
  key: 'app/cert-4j5vze02r/server.key'
principals:
  - username: 'test'
    password: 'be178c0543eb17f5f3043021c9e5fcf30285e557a4fc309cce97ff9ca6182912'
    permissions:
      - action: 'pub'
        permit: ['#']
      - action: 'sub'
        permit: ['#']

OpenEdge Function模块配置：
name: openedge_function
mark: modu-e1iluuach
hub:
  address: tcp://openedge_hub:1883
  username: test
  password: hahaha
rules:
  - id: rule-e1iluuac1
    subscribe:
      topic: data
      qos: 1
    compute:
      function: write
    publish:
      topic: data/tsdb
      qos: 1
functions:
  - name: 'write'
    runtime: 'python2.7'
    handler: 'write.handler'
    codedir: 'app/func-nyeosbbch'
    entry: "hub.baidubce.com/openedge-sandbox/openedge_function_runtime_python2.7:0.3.6"
    env:
      USER_ID: acuiot
    instance:
      min: 1
      max: 10
      timeout: 30s
```

通过上述配置不难发现，借助MQTTBOX向主题“data”发布消息，并通过“write”函数将该数据写入云端TSDB。

**需要说明的是**：为实际生产考虑，避免写入消息量过大时导致OpenEdge处理的消息过多，此处仅对写入失败进行错误信息提示，写入成功则不提示任何信息。

![通过MQTTBOX查看数据写入TSDB是否成功](../../images/develop/practice/tsdb/mqttbox-write-tsdb-success.png)

此外，也可以通过账户登录云端TSDB进行查看，具体如下:

![通过云端TSDB查看数据是否写入成功](../../images/develop/practice/tsdb/tsdb-check.png)

从上图不难看出，数据已经成功写入TSDB。