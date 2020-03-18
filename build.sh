#!/bin/sh
#
export CGO_ENABLED=1
export GOOS=windows
export GOARCH=amd64
export MINGW=x86_64-w64-mingw32
export CC=${MINGW}-gcc
export CXX=${MINGW}-g++
export DOCKER_HOST=127.0.0.1:2375

make rebuild image

