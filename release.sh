#!/bin/sh
#
set -e

go_version=$(go version | awk '{print substr($3, 3)}')
echo "Current golang version is $go_version, the minimum version of go required is 1.10.0"

# Use `godep restore ./...` checkout dependencies

find $GOPATH/src/github.com/docker -path '*/vendor' -type d | xargs -IX rm -r X

if [ -d output ]; then
    rm -rf output/*
fi
mkdir -p output

build_posix() {
    TAG=$1
    GOOS=$2
    GOARCH=$3
    echo "\nBuild $TAG with GOOS=$GOOS GOARCH=$GOARCH"

    if [ -d output/$TAG ]; then
        rm -rf output/$TAG
    fi
    mkdir -p output/$TAG

    tar cf - -C example . | tar xf - -C output/$TAG

    # engine
    env GOOS=$GOOS GOARCH=$GOARCH go build -o output/$TAG/native/bin/openedge ./cmd
    env GOOS=$GOOS GOARCH=$GOARCH go build -o output/$TAG/docker/bin/openedge ./cmd

    # modules
    env GOOS=$GOOS GOARCH=$GOARCH go build -o output/$TAG/native/bin/openedge_hub ./module/hub/cmd
    env GOOS=$GOOS GOARCH=$GOARCH go build -o output/$TAG/native/bin/openedge_function ./module/function/cmd
    env GOOS=$GOOS GOARCH=$GOARCH go build -o output/$TAG/native/bin/openedge_remote_mqtt ./module/remote/mqtt

    # function runtime python2.7
    cp module/function/runtime/python2.7/*.py  output/$TAG/native/bin/
    chmod +x output/$TAG/native/bin/openedge_function_runtime_python2.7.py

    # testing tools
    env GOOS=$GOOS GOARCH=$GOARCH go build -o output/$TAG/test/openedge_pubsub ./tools/pubsub
    env GOOS=$GOOS GOARCH=$GOARCH go build -o output/$TAG/test/openedge_benchmark ./tools/benchmark
    env GOOS=$GOOS GOARCH=$GOARCH go build -o output/$TAG/test/openedge_consistency ./tools/consistency

    cd output/$TAG/
    tar czvf ../openedge-$TAG-$VERSION.tar.gz *
    cd ../../
}

build_windows() {
    TAG=$1
    GOOS=$2
    GOARCH=$3
    CGOCC=$4
    echo "\nBuild $TAG with GOOS=$GOOS GOARCH=$GOARCH CGOCC=$CGOCC"

    if [ -d output/$TAG ]; then
        rm -rf output/$TAG
    fi
    mkdir -p output/$TAG

    tar cf - -C example . | tar xf - -C output/$TAG

    # engine
    env GOOS=$GOOS GOARCH=$GOARCH CC=$CGOCC go build -o output/$TAG/native/bin/openedge.exe ./cmd
    env GOOS=$GOOS GOARCH=$GOARCH CC=$CGOCC go build -o output/$TAG/docker/bin/openedge.exe ./cmd

    # modules
    env GOOS=$GOOS GOARCH=$GOARCH CC=$CGOCC go build -o output/$TAG/native/bin/openedge_hub.exe ./module/hub/cmd
    env GOOS=$GOOS GOARCH=$GOARCH CC=$CGOCC go build -o output/$TAG/native/bin/openedge_function.exe ./module/function/cmd
    env GOOS=$GOOS GOARCH=$GOARCH CC=$CGOCC go build -o output/$TAG/native/bin/openedge_remote_mqtt.exe ./module/remote/mqtt

    # function runtime python2.7
    cp module/function/runtime/python2.7/*.py  output/$TAG/native/bin/
    chmod +x output/$TAG/native/bin/openedge_function_runtime_python2.7.py

    # testing tools
    env GOOS=$GOOS GOARCH=$GOARCH CC=$CGOCC go build -o output/$TAG/test/openedge_pubsub.exe ./tools/pubsub
    env GOOS=$GOOS GOARCH=$GOARCH CC=$CGOCC go build -o output/$TAG/test/openedge_benchmark.exe ./tools/benchmark
    env GOOS=$GOOS GOARCH=$GOARCH CC=$CGOCC go build -o output/$TAG/test/openedge_consistency.exe ./tools/consistency

    cd output/$TAG/
    tar czvf ../openedge-$TAG-$VERSION.tar.gz *
    cd ../../
}

# build linux
build_posix     "linux-x86"       "linux"     "386"     "cc"
build_posix     "linux-x86_64"    "linux"     "amd64"   "cc"
build_posix     "linux-arm"       "linux"     "arm"     "cc"
build_posix     "linux-aarch64"   "linux"     "arm64"   "cc"

# build darwin
build_posix     "macos-x86_64"    "darwin"    "amd64"   "cc"

# build windows, please install `mingw` first
build_windows   "windows-x86"     "windows"   "386"     "i686-w64-mingw32-gcc"
build_windows   "windows-x86_64"  "windows"   "amd64"   "x86_64-w64-mingw32-gcc"
