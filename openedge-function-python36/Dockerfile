FROM python:3.6-alpine as base

FROM base as builder
RUN apk add --update --no-cache \
    gcc \
    linux-headers \
    make \
    musl-dev \
    python-dev \
    g++
RUN python -m pip install --upgrade pip
RUN pip install --install-option="--prefix=/install" pyyaml protobuf grpcio

FROM base
RUN apk add --update --no-cache gcc
COPY --from=builder /install /usr/local
WORKDIR /
COPY *.py /bin/
RUN chmod +x /bin/openedge-function-python36.py
ENTRYPOINT ["openedge-function-python36.py"]
