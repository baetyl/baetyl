#!/bin/sh
#
set -e
export GOPATH=$(pwd)/../../../../../
export GOROOT=$(pwd)/../../../baidu/go-env/go1-10-3-linux-amd64/
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH

../../../baidu/god-env/god-v0-6-0-linux-amd64/bin/god save -v ./...
find $GOPATH/src/github.com/docker -path '*/vendor' -type d | xargs -IX rm -r X
