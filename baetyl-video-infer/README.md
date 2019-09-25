# baetyl-video-infer

The baetyl-video-infer module is an official module of [BAETYL](https://baetyl.io), which is used for capturing video frame and AI model inference. For capturing video frame, baetyl-video-infer module can support IP-network camera, USB camera and video files. And the supported DL(Deep Learning) frameworks and layers are keeping in sync with OpenCV, more detailed can be found [here](https://github.com/opencv/opencv/wiki/Deep-Learning-in-OpenCV).

In addition, baetyl-video-infer module is highly integrated with other modules of BAETYL framework empowering the edge AI and some industry scenes. Besides, baetyl-video-infer also provides accelerated processing support for AI models based on CPU and specific hardware (such as OpenCL(OpenVINO, INFERENCE_ENGINE)).

## Build

Baetyl-video-infer module compiles depend on [GoCV](https://github.com/hybridgroup/gocv). Please make sure GoCV works properly before compiling.

```shell
go get github.com/baetyl/baetyl
cd $GOPATH/src/github.com/baetyl/baetyl
make rebuild # compile baetyl-video-infer
```

## Docker support

As above description, baetyl-video-infer module now can support CPU and specific hardware(Intel GPU, as known as OpenCL) for AI models inference. For the CPU, baetyl can support Linux-armv7, Linux-arm64, Linux-amd64, Linux-386, Darwin-amd64 platforms. And also provides accelerated processing support for AI models due to Intel GPU(OpenCL) of [OpenVINO framework](https://docs.openvinotoolkit.org/latest/index.html).

```shell
# CPU
docker build -t hub.baidubce.com/baetyl-bata/baetyl-gocv41:0.1.6-devel -f Dockerfile-gocv41-devel . # build devel-image for CPU
docker build -t hub.baidubce.com/baetyl-beta/baetyl-video-infer:0.1.6 -f Dockerfile . # build baetyl-video-infer image for CPU

# OpenVINO(OpenCL)
docker build -t hub.baidubce.com/baetyl-beta/baetyl-openvino:0.1.6-devel -f Dockerfile-openvino-devel . # build devel-image for OpenVINO
docker build -t hub.baidubce.com/baetyl-beta/baetyl-openvino:0.1.6-gocv41-devel -f Dockerfile-openvino-gocv41-devel . # build devel-image for OpenVINO and GoCV(OpenCV version is 4.1.0)
docker build -t hub.baidubce.com/baetyl-beta/baetyl-video-infer:0.1.6-openvino -f Dockerfile-openvino . # build baetyl-video-infer image of OpenVINO support
```

**NOTE**: 

- For OpenVINO(OpenCL), now only support Linux-amd64 platform;
- For CPU, cross compile is not support.

## How to use

Please refer to the example we provide.