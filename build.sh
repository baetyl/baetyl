#!/bin/sh
#
set -e

race=$1

# Uncomment the line below to checkout dependecies
# Use ../../../baidu/god-env/god-v0-6-0-linux-amd64/bin/god instead in Baidu intranet
#godep restore

echo "start to  build $race"
if [ -d output ]; then
    rm -rf output/*
fi
mkdir -p output

os=`go env | grep GOOS | awk -F"=" '{print $2}' | awk -F"\"" '{print $2}'`
echo "Current os: $os"

go_exe=`go env | grep GOEXE | awk -F"=" '{print $2}' | awk -F"\"" '{print $2}'`
echo "Current bin suffix: $go_exe"

go_version=$(go version | awk '{print substr($3, 3)}')
echo "Current golang version is $go_version, the minimum version of go required is 1.10.0"

find $GOPATH/src/github.com/docker -path '*/vendor' -type d | xargs -IX rm -r X

tar cf - -C example . | tar xf - -C output/

# engine
go build $race -o output/native/bin/openedge$go_exe ./cmd
go build $race -o output/docker/bin/openedge$go_exe ./cmd

# modules
go build $race -o output/native/bin/openedge_hub$go_exe ./module/hub/cmd
go build $race -o output/native/bin/openedge_function$go_exe ./module/function/cmd
go build $race -o output/native/bin/openedge_remote_mqtt$go_exe ./module/remote/mqtt

# function runtime python 2.7
cp ./module/function/runtime/python2.7/*.py  output/native/bin/
chmod +x output/native/bin/openedge_function_runtime_python2.7.py

go build $race -o output/test/pubsub$go_exe ./tools/pubsub
go build $race -o output/test/openedge_benchmark$go_exe ./tools/benchmark
go build $race -o output/test/openedge_consistency$go_exe ./tools/consistency
