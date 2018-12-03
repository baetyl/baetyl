#!/bin/sh
#
set -e
export GOPATH=$(pwd)/../../../../../
export GOROOT=$(pwd)/../../../baidu/go-env/go1-10-3-linux-amd64/
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH

../../../baidu/god-env/god-v0-6-0-linux-amd64/bin/god restore -v
find $GOPATH/src/github.com/docker -path '*/vendor' -type d | xargs -IX rm -r X

if [ -d output ]; then
    rm -rf output
fi
mkdir -p output
tar cf - -C example . | tar xf - -C output/$TAG

build_posix() {
    TAG=$1
    GOOS=$2
    GOARCH=$3
    echo "Build $TAG with GOOS=$GOOS GOARCH=$GOARCH"

    if [ -d output/$TAG ]; then
        rm -rf output/$TAG
    fi
    mkdir -p output/$TAG/bin

    # engine
    env GOOS=$GOOS GOARCH=$GOARCH go build -o output/$TAG/native/bin/openedge ./cmd
    env GOOS=$GOOS GOARCH=$GOARCH go build -o output/$TAG/docker/bin/openedge ./cmd
    # modules
    env GOOS=$GOOS GOARCH=$GOARCH go build -o output/$TAG/native/bin/openedge_hub ./module/hub/cmd
    env GOOS=$GOOS GOARCH=$GOARCH go build -o output/$TAG/native/bin/openedge_function ./module/function/cmd
    env GOOS=$GOOS GOARCH=$GOARCH go build -o output/$TAG/native/bin/openedge_remote_mqtt ./module/remote/mqtt
    cp module/function/runtime/python2.7/*.py  output/$TAG/bin/
    chmod +x output/$TAG/bin/openedge_function_runtime_python2.7.py
}

build_windows() {
    TAG=$1
    GOOS=$2
    GOARCH=$3
    CGOCC=$4
    echo "Build $TAG with GOOS=$GOOS GOARCH=$GOARCH CGOCC=$CGOCC"

    if [ -d output/$TAG ]; then
        rm -rf output/$TAG
    fi
    mkdir -p output/$TAG/bin
    mkdir -p output/$TAG/libexec
    env GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=1 CC=$CGOCC go build -o output/$TAG/bin/openedge.exe ./hub/cmd
    cp hub/runtime/packet.py  output/$TAG/libexec/packet.py
    cp hub/runtime/message_pb2.py  output/$TAG/libexec/message_pb2.py
    cp hub/runtime/openedge_python2.7_windows.py  output/$TAG/libexec/openedge_python2.7

    tar cf - -C example . | tar xf - -C output/$TAG
    env GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=1 CC=$CGOCC go build -o output/$TAG/test/openedge_benchmark.exe ./tools/benchmark
    env GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=1 CC=$CGOCC go build -o output/$TAG/test/openedge_consistency.exe ./tools/consistency
}

build_posix     "linux-x86_64"    "linux"     "amd64" "cc"
build_posix     "linux-arm"       "linux"     "arm"   "cc"
