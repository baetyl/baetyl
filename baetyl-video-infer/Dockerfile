FROM hub.baidubce.com/baetyl-beta/baetyl-gocv41:0.1.5-devel as devel
FROM ubuntu:18.04

RUN apt-get update -y && \
    apt-get upgrade -y && \
    apt-get -y --no-install-recommends install pkg-config libgtk2.0-dev libavcodec-dev libavformat-dev libswscale-dev libtbb2 libtbb-dev libjpeg-dev libpng-dev libtiff-dev libdc1394-22-dev


# OpenCV 4.1.0 shared objects from devel
COPY --from=devel /usr/local/lib /usr/local/lib
COPY --from=devel /usr/local/lib/pkgconfig/opencv4.pc /usr/local/lib/pkgconfig/opencv4.pc
COPY --from=devel /usr/local/include/opencv4/opencv2 /usr/local/include/opencv4/opencv2

ENV PKG_CONFIG_PATH /usr/local/lib/pkgconfig
ENV LD_LIBRARY_PATH /usr/local/lib

COPY baetyl-video-infer /bin/
ENTRYPOINT ["baetyl-video-infer"]