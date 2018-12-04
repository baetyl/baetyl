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

version_info="package version\n\n// Version the version of this binary\nconst Version = \"$version\""
echo $version_info > version/version.go

go_version=`go version | awk '{print substr($3, 3)}'`
echo "Current golang version is $go_version, the minimum version of go required is 1.10.0"

# Use `godep restore ./...` checkout dependencies

find $GOPATH/src/github.com/docker -path '*/vendor' -type d | xargs -IX rm -r X

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

    if [ $GOOS == "windows" ];then
        GOEXE=".exe"
    else
        GOEXE=""
    fi

    if [ -d output/$TAG ]; then
        rm -rf output/$TAG
    fi
    mkdir -p output/$TAG

    tar cf - -C example/native . | tar xf - -C output/$TAG

    # engine
    env GOOS=$GOOS GOARCH=$GOARCH go build -o output/$TAG/bin/openedge$GOEXE ./cmd

    # modules
    env GOOS=$GOOS GOARCH=$GOARCH go build -o output/$TAG/bin/openedge_hub$GOEXE ./module/hub/cmd
    env GOOS=$GOOS GOARCH=$GOARCH go build -o output/$TAG/bin/openedge_function$GOEXE ./module/function/cmd
    env GOOS=$GOOS GOARCH=$GOARCH go build -o output/$TAG/bin/openedge_remote_mqtt$GOEXE ./module/remote/mqtt

    # function runtime python2.7
    cp module/function/runtime/python2.7/*.py  output/$TAG/bin/
    chmod +x output/$TAG/bin/openedge_function_runtime_python2.7.py

    cd output/$TAG/
    tar czvf ../openedge-$TAG-$version.tar.gz *
    cd ../../
}

# build for multi CPU && os, if for windows, please install `mingw` first
build   "linux-x86"       "linux"     "386"     "cc"   
build   "linux-x86_64"    "linux"     "amd64"   "cc"     
build   "linux-arm"       "linux"     "arm"     "cc"
build   "linux-aarch64"   "linux"     "arm64"   "cc"
build   "darwin-x86_64"   "darwin"    "amd64"   "cc"
build   "windows-x86"     "windows"   "386"     "i686-w64-mingw32-gcc"
build   "windows-x86_64"  "windows"   "amd64"   "x86_64-w64-mingw32-gcc"
