# How to import third-party libraries for Python runtime

**Statement**

- The operating system as mentioned in this document is Ubuntu18.04.
- The version of runtime is Python3.6, and for Python2.7, configurations are the same except for the language differences when coding the scripts.
- The MQTT client toolkit as mentioned in this document is [MQTTBOX](../Resources-download.md#mqttbox-download).
- In this document, the third-party libraries we'll import are [`requests`](https://pypi.org/project/requests) and [`Pytorch`](https://pytorch.org/).
- In this article, the service created based on the Hub module is called `localhub` service. And for the test case mentioned here, the `localhub` service, function calculation service, and other services are configured as follows:

```yaml
# The configuration of localhub service
# Configuration file location is: var/db/baetyl/localhub-conf/service.yml
listen:
  - tcp://0.0.0.0:1883
principals:
  - username: 'test'
    password: 'hahaha'
    permissions:
      - action: 'pub'
        permit: ['#']
      - action: 'sub'
        permit: ['#']

# The configuration of Local Function Manager service
# Configuration file location is: var/db/baetyl/function-manager-conf/service.yml
hub:
  address: tcp://localhub:1883
  username: test
  password: hahaha
rules:
  - clientid: localfunc-1
    subscribe:
      topic: py
    function:
      name: sayhi3
    publish:
      topic: py/hi
functions:
  - name: sayhi3
    service: function-sayhi3
    instance:
      min: 0
      max: 10
      idletime: 1m

# The configuration of application.yml
# Configuration file location is: var/db/baetyl/application.yml
version: v0
services:
  - name: localhub
    image: baetyl-hub
    replica: 1
    ports:
      - 1883:1883
    mounts:
      - name: localhub-conf
        path: etc/baetyl
        readonly: true
      - name: localhub-data
        path: var/db/baetyl/data
      - name: localhub-log
        path: var/log/baetyl
  - name: function-manager
    image: baetyl-function-manager
    replica: 1
    mounts:
      - name: function-manager-conf
        path: etc/baetyl
        readonly: true
      - name: function-manager-log
        path: var/log/baetyl
  - name: function-sayhi3
    image: baetyl-function-python36
    replica: 0
    mounts:
      - name: function-sayhi-conf
        path: etc/baetyl
        readonly: true
      - name: function-sayhi-code
        path: var/db/baetyl/function-sayhi
        readonly: true
volumes:
  # hub
  - name: localhub-conf
    path: var/db/baetyl/localhub-conf
  - name: localhub-data
    path: var/db/baetyl/localhub-data
  - name: localhub-log
    path: var/db/baetyl/localhub-log
  # function manager
  - name: function-manager-conf
    path: var/db/baetyl/function-manager-conf
  - name: function-manager-log
    path: var/db/baetyl/function-manager-log
  # function python runtime sayhi
  - name: function-sayhi-conf
    path: var/db/baetyl/function-sayhi-conf
  - name: function-sayhi-code
    path: var/db/baetyl/function-sayhi-code
```

Generally, using the Python Standard Library may not meet our needs. In fact, it is often necessary to import some third-party libraries. Two examples are given below.

## Import `requests` third-party libraries

Suppose we want to crawl a website and get the response. Here, we can import a third-party library [`requests`](https://pypi.org/project/requests). How to import it, as shown below:

- Step 1: change path to the directory of Python scripts, then download `requests` package and its dependency packages(idna、urllib3、chardet、certifi)

```shell
cd /directory/of/Python/script
pip download requests
```

- Step 2: inflate the downloaded `.whl` files for getting the source packages, then remove useless `.whl` files and package-description files

```shell
unzip -d . "*.whl"
rm -rf *.whl *.dist-info
```

- Step 3: make the current directory be a package

```shell
touch __init__.py
```

- Step 4: import the third-party library `requests` in the Python script as shown below:

```python
import requests
```

- Step 5: execute your Python script

```shell
python your_script.py
```

If the above operations are normal, the resulting script directory structure is as shown in the following figure.

![the directory of the Python script](../../images/customize/python-third-lib-dir-requests.png)

Now we write the Python script `get.py` to get the headers information of [https://baetyl.io](https://baetyl.io), assuming the trigger condition is that Python3.6 runtime receives the "A" command from the `localhub` service. More detailed contents are as follows:

```python
#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import requests

def handler(event, context):
    """
    data: {"action": "A"}
    """
    if 'action' in event:
        if event['action'] == 'A':
            r = requests.get('https://baetyl.io')
            if str(r.status_code) == '200':
                event['info'] = dict(r.headers)
            else:
                event['info'] = 'exception found'
        else:
            event['info'] = 'action error'
    else:
        event['error'] = 'action not found'

return event
```

The configuration of Python function runtime is as below:

```yaml
# The configuration of Python function runtime
functions:
  - name: 'sayhi3'
    handler: 'get.handler'
    codedir: 'var/db/baetyl/function-sayhi'
```

As above, after receiving the message publish to the topic `py`, the `localhub` service will call the `get.py` script to handle, and following it publish the result to the topic `py/hi`. So in the test case, we use MQTTBOX to subscribe the topic `py/hi` and publish the message `{"action": "A"}` to the `localhub` service by the topic `py`. If everything works correctly, MQTTBOX can receive the message of the topic `py/hi` which contains the headers information of [https://baetyl.io](https://baetyl.io) as shown below.

![Get the header information of https://baetyl.io](../../images/customize/write-python-script-third-lib-requests.png)

## Import `Pytorch` third-party libraries

`Pytorch` is a widely used deep learning framework for machine learning. We can import a third-party library [`Pytorch`](https://pytorch.org/) to use its functions. How to import it, as shown below:

- Step 1: change path to the directory of Python scripts, then download `Pytorch` package and its dependency packages(PIL、caffee2、numpy、six、torchvision)

```shell
cd /directory/of/Python/script
pip3 download torch torchvision
```

- Step 2: inflate the downloaded `.whl` files for getting the source packages, then remove useless `.whl` files and package-description files

```shell
unzip -d . *.whl
rm -rf *.whl *.dist-info
```

- Step 3: make the current directory be a package

```shell
touch __init__.py
```

- Step 4: import the third-party library `Pytorch` in the Python script as shown below:

```python
import torch
```

- Step 5: execute your Python script

```shell
python your_script.py
```

If the above operations are normal, the resulting script directory structure is as shown in the following figure.

![the directory of the Python script](../../images/customize/python-third-lib-dir-Pytorch.png)

Now we write the Python script `calc.py` to use functions provided by `Pytorch` for generating a random tensor, assuming the trigger condition is that Python3.6 runtime receives the "B" command from the `localhub` service. More detailed contents are as follows:

```python
#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import torch

def handler(event, context):
  """
  data: {"action": "B"}
  """
  if 'action' in event:
    if event['action'] == 'B':
      x = torch.rand(5, 3)
      event['info'] = x.tolist()
    else:
      event['info'] = 'exception found'
  else:
    event['error'] = 'action not found'

  return event
```

The configuration of Python function runtime is as below:

```yaml
# The configuration of Python function runtime
functions:
  - name: 'sayhi3'
    handler: 'calc.handler'
    codedir: 'var/db/baetyl/function-sayhi'
```

As above, after receiving the message publish to the topic `py`, the `localhub` service will call the `calc.py` script to handle, and following it publish the result to the topic `py/hi`. So in the test case, we use MQTTBOX to subscribe the topic `py/hi` and publish the message `{"action": "B"}` to the `localhub` service by the topic `py`. If everything works correctly, MQTTBOX can receive the message of the topic `py/hi` in which we can get a random tensor as shown below.

![generate a random tensor](../../images/customize/write-python-script-third-lib-Pytorch.png)