FROM ubuntu:16.04 as base

RUN apt-get update \
        && apt-get install -y software-properties-common --no-install-recommends \
        && add-apt-repository -y ppa:jonathonf/python-3.6 \
        && apt update \
        && apt install -y python3.6 \
        && ln -s /usr/bin/python3.6 /usr/bin/python \
        && rm /usr/bin/python3 \
        && ln -s /usr/bin/python3.6 /usr/bin/python3

FROM base as builder
RUN apt-get install -y gcc g++ python3.6-dev python3-setuptools python3-pip --no-install-recommends \
    && ln -s /usr/bin/pip3 /usr/bin/pip \
    && python -m pip install --upgrade pip \
        && pip install --install-option="--prefix=/install" pyyaml protobuf grpcio

FROM base
COPY --from=builder /install/lib/python3.6/site-packages /usr/local/lib/python3.6/dist-packages/
RUN rm -rf /var/lib/apt/lists/*
