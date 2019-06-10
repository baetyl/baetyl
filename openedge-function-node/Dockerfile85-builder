FROM ubuntu:16.04 as base

FROM base as builder
RUN apt-get update \
        && apt-get install -y build-essential wget python \
        && wget https://nodejs.org/dist/v8.5.0/node-v8.5.0.tar.gz \
        && tar vxzf node-v8.5.0.tar.gz \
        && cd node-v8.5.0 \
        && ./configure --prefix=__package \
        && make \
        && make install


FROM base
COPY --from=builder node-v8.5.0/__package /usr/local
RUN rm -rf /var/lib/apt/lists/*

