# baetyl-video-infer:0.1.5

```shell
# make devel image
cd $GOPATH/src/icode.baidu.com/baidu/bce-iot/edge-projects/baetyl-video-infer
docker build -t hub.baidubce.com/baetyl-beta/baetyl-gocv41:0.1.5-devel -f Dockerfile-gocv41-devel .
# make module image
docker run -it -v $GOPATH:/root/go hub.baidubce.com/baetyl-beta/baetyl-gocv41:0.1.5-devel
cd $GOPATH/src/icode.baidu.com/baidu/bce-iot/edge-projects/baetyl-video-infer/
make rebuild
exit
docker build -t hub.baidubce.com/baetyl-beta/baetyl-video-infer:0.1.5 -f Dockerfile .
```

# baetyl-video-infer:0.1.5-openvino
```shell
# baetyl-openvino:0.1.5-devel
cd $GOPATH/src/icode.baidu.com/baidu/bce-iot/edge-projects/baetyl-video-infer
docker build -t hub.baidubce.com/baetyl-beta/baetyl-openvino:0.1.5-devel -f Dockerfile-openvino-devel .
# baetyl-openvino:0.1.5-gocv41-devel
cd $GOPATH/src/icode.baidu.com/baidu/bce-iot/edge-projects/baetyl-video-infer
docker build -t hub.baidubce.com/baetyl-beta/baetyl-openvino:0.1.5-gocv41-devel -f Dockerfile-openvino-gocv41-devel .
# baetyl-video-infer:0.1.5-openvino
docker run -it -v ${GOPATH}:/root/go hub.baidubce.com/baetyl-beta/baetyl-openvino:0.1.5-gocv41-devel
cd $GOPATH/src/icode.baidu.com/baidu/bce-iot/edge-projects/baetyl-video-infer/
go build -tags openvino .
exit
docker build -t hub.baidubce.com/baetyl-beta/baetyl-video-infer:0.1.5-openvino -f Dockerfile-openvino .
```