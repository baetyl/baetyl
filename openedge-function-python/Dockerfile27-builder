FROM ubuntu:16.04 as base

RUN apt-get update \
        && apt-get install -y python2.7 --no-install-recommends \
        && ln -s /usr/bin/python2.7 /usr/bin/python

FROM base as builder
RUN apt-get update \
        && apt-get install -y python-pip \
        && pip install --install-option="--prefix=/install" pyyaml protobuf grpcio

FROM base
COPY --from=builder /install/lib/python2.7/site-packages /usr/local/lib/python2.7/dist-packages
RUN rm -rf /var/lib/apt/lists/*
