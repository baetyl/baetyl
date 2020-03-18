#!/bin/sh
#
export CGO_ENABLED=1
export GOOS=linux
export GOARCH=amd64
#export MINGW=/usr/bin/x86_64-w64-mingw32
#export CC=${MINGW}-gcc
#export CXX=${MINGW}-g++
#export DOCKER_HOST=127.0.0.1:2375

#cp -r example/docker/var /mnt/c/baetyl/; exit
make rebuild image
#cp output/${GOOS}/${GOARCH}/baetyl/bin/baetyl /mnt/c/baetyl/baetyl.exe
#mkdir -p tmphub
#cp output/${GOOS}/${GOARCH}/baetyl-hub/bin/baetyl-hub tmphub/baetyl-hub
#cd tmphub && docker build -t baetyl-hub:latest .
