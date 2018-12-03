FROM python:2.7-alpine as base

FROM base as builder
RUN apk add --update --no-cache \
    gcc \
    linux-headers \
    make \
    musl-dev \
    python-dev \
    g++
ENV GRPC_PYTHON_VERSION 1.15.0
RUN python -m pip install --upgrade pip
RUN pip install --install-option="--prefix=/install" pyyaml protobuf grpcio==${GRPC_PYTHON_VERSION}

FROM base
RUN apk add --update --no-cache gcc
COPY --from=builder /install /usr/local
WORKDIR /home/openedge
COPY output/native/bin/*.py ./bin/
RUN chmod +x ./bin/openedge_function_runtime_python2.7.py
ENTRYPOINT ["bin/openedge_function_runtime_python2.7.py"]
CMD []
