# Image capturing and AI model inference with Video infer Service

**Statement**

- In this document, the used operating system is Ubuntu18.04
- In this document, the selected camera is Hikvision DS-IPC-B12-1
- In this document, image cpaturing and AI model inference both depend on CPU with video infer service
- In this document, the selected AI model is [ssd_mobilenet_v1_coco_2017_11_17](https://baetyl.cdn.bcebos.com/example/guides/ssd_mobilenet_v1_coco_2017_11_17.tar.gz)
- In this document, the MQTT client toolkit which is used to connect to baetyl-hub service is [MQTTBox](../Resources.html#mqttbox-download)
- In this document, the connection method between other service and baetyl-hub service is TCP

## Workflow

- Step 1: Install Baetyl on Ubuntu18.04, more detailed contents please refer to [Qick-install-Baetyl](../install/Quick-Install.md)
- Step 2: Write all services configuration file, and start Baetyl with command `sudo systemctl start baetyl`. Also, we can view the Baetyl's running status and all running containers through command `sudo systemctl status baetyl` and `docker ps`. More detailed contents of all services configuration please refer to [Configuration](#configuration)
- Step 3: Select MQTTBox as the MQTT client that connect to BAETYL framework, more detailed contents please refer to [Device connect to BAETYL framework with Hub Service](Device-connect-to-BAETYL-framework-with-hub-service.md)
- Step 4: Subscribe topic `video/infer/result` and observe whether it can be received normally

The workflow of above operation is as follow:

![Image capturing and AI model inference with Video infer Service](../images/guides/inference/video-infer-flow.png)

## Configuration

The application configuration of Baetyl is located in `var/db/baetyl/application.yml`, more detailed contents are as follows:

```yaml
version: V0
services:
  - name: localhub
    image: 'hub.baidubce.com/baetyl/baetyl-hub:latest'
    replica: 1
    ports:
      - '1883:1883'
    mounts:
      - name: localhub-conf
        path: etc/baetyl
        readonly: true
      - name: localhub-persist-data
        path: var/db/baetyl/data
      - name: demo-log
        path: var/log/baetyl
  - name: function-manager
    image: 'hub.baidubce.com/baetyl/baetyl-function-manager:latest'
    replica: 1
    mounts:
      - name: function-manager-conf
        path: etc/baetyl
        readonly: true
      - name: demo-log
        path: var/log/baetyl
  - name: function-python
    image: 'hub.baidubce.com/baetyl/baetyl-function-python36:0.1.6-opencv41'
    replica: 0
    mounts:
      - name: function-python-conf
        path: etc/baetyl
        readonly: true
      - name: function-python-code
        path: var/db/baetyl/code
        readonly: true
      - name: image-data
        path: var/db/baetyl/image
  - name: video-infer
    image: 'hub.baidubce.com/baetyl-beta/baetyl-video-infer:latest'
    replica: 1
    mounts:
      - name: infer-person-model
        path: var/db/baetyl/model
        readonly: true
      - name: image-data
        path: var/db/baetyl/image
      - name: demo-log
        path: var/log/baetyl
      - name: video-infer-conf
        path: etc/baetyl
        readonly: true
volumes:
  - name: localhub-conf
    path: var/db/baetyl/localhub-conf
  - name: localhub-persist-data
    path: var/db/baetyl/localhub-persist-data
  - name: demo-log
    path: var/db/baetyl/demo-log
  - name: function-manager-conf
    path: var/db/baetyl/function-manager-conf
  - name: function-python-conf
    path: var/db/baetyl/function-python-conf
  - name: function-python-code
    path: var/db/baetyl/function-python-code
  - name: image-data
    path: var/db/baetyl/image-data
  - name: remote-mqtt-conf
    path: var/db/baetyl/remote-mqtt-conf
  - name: infer-person-model
    path: var/db/baetyl/infer-person-model
  - name: video-infer-conf
    path: var/db/baetyl/video-infer-conf
```

The configuration of baetyl-hub service is located in `var/db/baetyl/localhub-conf/service.yml`, more detailed contents are as follows:

```yaml
listen:
  - tcp://0.0.0.0:1883
principals:
  - username: test
    password: hahaha
    permissions:
      - action: 'pub'
        permit: ['#']
      - action: 'sub'
        permit: ['#']
logger:
  path: var/log/baetyl/localhub-service.log
  level: "debug"
```

The configuration of baetyl-function-manager service is located in `var/db/baetyl/function-manager-conf/service.yml`, more detailed contents are as follows:

```yaml
server:
  address: 0.0.0.0:50051
hub:
  address: tcp://localhub:1883
  username: test
  password: hahaha
functions:
  - name: analyse
    service: function-python
    instance:
      max: 10
logger:
  path: var/log/baetyl/func-service.log
  level: "debug"
```

The configuration of baetyl-function-python service is loacated in `var/db/baetyl/function-python-conf/service.yml`, more detailed contents are as follows:

```yaml
functions:
  - name: 'analyse'
    handler: 'analyse.handler'
    codedir: 'var/db/baetyl/code'
logger:
  path: "var/log/baetyl/python-service.log"
  level: "debug"
```

The configuration of baetyl-video-infer service is located in `var/db/baetyl/video-infer-conf/service.yml`, more detailed contents are as follows:

```yaml
hub:
  address: tcp://localhub:1883
  username: test
  password: hahaha
video:
  uri: "rtsp://admin:admin@192.168.1.2:554/Streaming/channels/1/"
  limit:
    fps: 1
infer:
  model: var/db/baetyl/model/frozen_inference_graph.pb
  config: var/db/baetyl/model/ssd_mobilenet_v1_coco_2017_11_17.pbtxt
process:
  before:
    swaprb: true
    width: 300
    hight: 300
  after:
    function:
      name: analyse
functions:
  - name: analyse
    address: function-manager:50051
logger:
  path: var/log/baetyl/infer-service.log
  level: "debug"
```

**NOTE**: 

- The `uri` configuration item stands for the address of IP-camera, it's normal format is `rtsp://<username>:<password>@<ip>:<port>/Streaming/channels/<stream_number>`
    - `<username>` and `<password>` are the login authentication element
    - `<ip>` is the IP-address of camera
    - `<port>` is the port number of RTSP protocol, the default value is `554`
    - `<stream_number>` is the channel number, if it is equal to `1`, it indicates that the main stream is being captured; if it is equal to `2`, it indicates that the secondary stream is being captured

Besides, video infer service can also support to capture frame from USB-camera and video file. In addition, if the captured device is USB-camera, the `uri` configuration item is deviceID, the normal value of it is `0`, and also need to map the device `/dev/video0` into the container. If the captured device is an video file, the `uri` configuration item is the path of the video file, and also need to mount it as volume on video infer service. More detailed contents about volume create and mount on service please refer to [How to correctly create and mount volume on service](https://cloud.baidu.com/doc/BIE/s/Cjzdn8xig).

For example, the configuration of USB-camera are as follows:

```yaml
video:
  uri: "0"
  limit:
    fps: 1
```

And it also need to map device `/dev/video0` to video infer service, more detailed contents are as follows:

```yaml
version: V0
services:
  - name: video-infer
    image: 'hub.baidubce.com/baetyl-beta/baetyl-video-infer:latest'
    replica: 1
    devices:
      - /dev/video0 # map the device `/dev/video0` to container
    mounts:
      - name: infer-person-model
        path: var/db/baetyl/model
        readonly: true
      - name: image-data
        path: var/db/baetyl/image
      - name: demo-log
        path: var/log/baetyl
      - name: video-infer-conf
        path: etc/baetyl
        readonly: true
```

It is not difficult to find that the above mapping method also applies to serial port device and other mounted devices.

## Test and verification

As mentioned in the beginning of this document, we select an object detection model of [tensorflow](https://www.tensorflow.org) framework(`ssd_mobilenet_v1_coco_2017_11_17`) which can be used to detection person, banner, apple and other fruits. Here, we provide an example python script(be used in baetyl-function-python service) of person detection, more detailed contents are as follows:

```python
#!/usr/bin/env python
# -*- coding:utf-8 -*-
"""
function to analyse video infer result in python
"""
import numpy as np

location = "var/db/baetyl/image/{}.jpg"
classes = {
    1: 'person'
}

def handler(event, context):
    """
    function handler
    """
    data = np.fromstring(event, np.float32)
    mat = np.reshape(data, (-1, 7))
    objects = []
    scores = {}
    for obj in mat:
        clazz = int(obj[1])
        if clazz in classes:
            score = float(obj[2])
            if classes[clazz] not in scores or scores[classes[clazz]] < score:
                scores[classes[clazz]] = score
            if score < 0.6:
                continue
            objects.append({
                'class': classes[clazz],
                'score': score,
                'left': float(obj[3]),
                'top': float(obj[4]),
                'right': float(obj[5]),
                'bottom': float(obj[6])
            })

    res = {}
    res["imageDiscard"] = len(objects) == 0
    res["imageObjects"] = objects
    res["imageScores"] = scores
    res["publishTopic"] = "video/infer/result"
    res["messageTimestamp"] = int(context["messageTimestamp"]/1000000)
    if len(objects) != 0:
        res["imageLocation"] = location.format(context["messageTimestamp"])

    return res
```

If you want to detect other supported object, modify it directly. The supported detection object please refer to [mscoco_label_map](https://github.com/tensorflow/models/blob/master/research/object_detection/data/mscoco_label_map.pbtxt).

When everything is done, we start Baetyl and view the baetyl running status and running containers through the command `sudo systemctl status baetyl` and `docker ps`.

![the running status of baetyl](../images/guides/inference/running-status-of-baetyl.png)

![the running containers](../images/guides/inference/running-container-view.png)

We can find the Baetyl service is running, and the running containers include: `baetyl-hub`, `baetyl-function-manager`, `baetyl-function-python` and `baetyl-video-infer`.

Then we start MQTTBox and set the connection configuration. If it works normally, we can find MQTTBox is connecting to baetyl-hub service.

![ MQTTBox is connection to baetyl-hub service](../images/guides/inference/MQTTBox-connect-to-hub.png)

As mentioned in workflow, we subscribe topic `video/infer/result` through MQTTBox, and observe the message receiving result.

![MQTTBox receive message, none person](../images/guides/inference/model-infer-result-none-person.png)

![MQTTBox receive message, person](../images/guides/inference/model-infer-result-person.png)

As above shown, one image shows thant it has detected person, and the other one has not detected person. Compare with them, we can find the message has `class` item when person detected. Besides, it also shows the location information of the detected person. That is to say we can mark the edge of the detected person with it. An example python script was provided in [here](https://github.com/baetyl/baetyl/blob/master/baetyl-video-infer/example/docker/var/db/baetyl/function-python-code/mark.py).

In summary, we have implemented IP-camera frame capturing and AI model inference based on the video infer service.