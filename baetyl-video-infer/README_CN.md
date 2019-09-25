# baetyl-video-infer

baetyl-video-infer 模块是 [BAETYL](https://baetyl.io) 框架新推出的集视频抽帧和 AI 模型推断于一体的模块，视频采集方面能够支持对 IP 网络摄像头、USB 摄像头及视频文件进行抽帧，AI 模型推断方面支持力度与 [OpenCV DNN 模块支持模型与网络](https://github.com/opencv/opencv/wiki/Deep-Learning-in-OpenCV) 保持一致。

另外，baetyl-video-infer 模块与 BAETYL 框架其他模块高度融合，整体上为边缘 AI 和相关行业场景赋能。特别地，baetyl-video-infer 模块提供对基于 CPU 和特定硬件（目前仅支持 Intel GPU，即 OpenCL）对 AI 模型推断进行加速处理。

## 源码编译

baetyl-video-infer 编译依赖 [GoCV](https://github.com/hybridgroup/gocv)，编译 baetyl-video-infer 前请确保 GoCV 已能够正常使用。

```shell
go get github.com/baetyl/baetyl
cd $GOPATH/src/github.com/baetyl/baetyl
make rebuild # 编译可执行程序 baetyl-video-infer
```

## Docker 支持

针对 Docker 容器支持，baetyl-video-infer 现已能够支持基于 CPU 支持 Linux-386、Linux-armv7、Linux-arm64、Linux-amd64 及 Darwin-amd64（实际运行 Linux-amd64 平台镜像）等平台；另外，还能够基于 OpenCL 通过 OpenVINO 对 AI 模型推断作加速处理，关于 OpenVINO 的更多内容可以参考 [官方文档](https://docs.openvinotoolkit.org/latest/index.html)。

```shell
# CPU
docker build -t hub.baidubce.com/baetyl-bata/baetyl-gocv41:0.1.6-devel -f Dockerfile-gocv41-devel . # 打包、构建 CPU 版基础镜像
docker build -t hub.baidubce.com/baetyl-beta/baetyl-video-infer:0.1.6 -f Dockerfile . # 打包、构建 CPU 版 baetyl-video-infer 镜像

# OpenVINO(OpenCL)
docker build -t hub.baidubce.com/baetyl-beta/baetyl-openvino:0.1.6-devel -f Dockerfile-openvino-devel . # 打包构建 OpenVINO 基础镜像
docker build -t hub.baidubce.com/baetyl-beta/baetyl-openvino:0.1.6-gocv41-devel -f Dockerfile-openvino-gocv41-devel . # 打包构建 OpenVINO 和 GoCV41(OpenCV 版本 4.1.0) 基础镜像
docker build -t hub.baidubce.com/baetyl-beta/baetyl-video-infer:0.1.6-openvino -f Dockerfile-openvino . # 打包、构建 OpenVINO(OpenCL) 版 baetyl-video-infer 镜像
```

**注意**：

- 针对 OpenVINO(OpenCL)，现在仅支持 Linux-amd64 平台；
- 针对 CPU，不支持跨平台编译（底层依赖 CGo）。

## 更多应用

请参考我们提供的 example 中的例子。