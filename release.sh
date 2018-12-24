#!/bin/sh
#
set -e

while [ -n "$1" ]
do
	case "$1" in
		-v)version=$2
           shift ;;
	esac
	shift
done

version_info="package module\n\n// Version the version of this binary\nconst Version = \"$version\""
echo $version_info > module/version.go

go_version=`go version | awk '{print substr($3, 3)}'`
echo "Current golang version is $go_version, the minimum version of go required is 1.10.0"

if [ -d output ]; then
    rm -rf output/*
fi
mkdir -p output

build() {
    TAG=$1
    GOOS=$2
    GOARCH=$3
    CGOCC=$4
    echo "\nBuild $TAG with GOOS=$GOOS GOARCH=$GOARCH CGOCC=$CGOCC"

    if [ $GOOS = "windows" ];then
        GOEXE=".exe"
    else
        GOEXE=""
    fi

    if [ -d output/$TAG ]; then
        rm -rf output/$TAG
    fi
    mkdir -p output/$TAG

    # docker release for multiple cpu && os
    tar cf - -C example/docker . | tar xf - -C output/$TAG

    # native release for multiple cpu && os(enable comment script below)
    # tar cf - -C example/native . | tar xf - -C output/$TAG

    # engine
    env GOOS=$GOOS GOARCH=$GOARCH go build -o output/$TAG/bin/openedge$GOEXE .

    # modules
    # env GOOS=$GOOS GOARCH=$GOARCH go build -o output/$TAG/bin/openedge-hub$GOEXE ./openedge-hub
    # env GOOS=$GOOS GOARCH=$GOARCH go build -o output/$TAG/bin/openedge-function$GOEXE ./openedge-function
    # env GOOS=$GOOS GOARCH=$GOARCH go build -o output/$TAG/bin/openedge-remote-mqtt$GOEXE ./openedge-remote-mqtt

    # function runtime python2.7
    # cp openedge-function-runtime-python27/*.py  output/$TAG/bin/
    # chmod +x output/$TAG/bin/openedge_function_runtime_python27.py

    cd output/$TAG/
    tar czvf ../openedge-$TAG-$version.tar.gz *
    cd ../../
}

# build for multiple cpu && os, if for windows, please install `mingw` first
build   "linux-x86"         "linux"     "386"     "cc"   
build   "linux-x86_64"      "linux"     "amd64"   "cc"  
build   "linux-armv7"       "linux"     "arm"     "cc"
build   "linux-aarch64"     "linux"     "arm64"   "cc"
build   "windows10-x86"     "windows"   "386"     "i686-w64-mingw32-gcc"
build   "windows10-x86_64"  "windows"   "amd64"   "x86_64-w64-mingw32-gcc"

# must be built on MacOS
# build   "darwin-x86_64"     "darwin"    "amd64"   "cc"
