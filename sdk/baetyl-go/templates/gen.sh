#!/bin/sh

MODULE=$1

sed "s/{{.MODULE}}/${MODULE}/g" ./templates/Makefile > ../../${MODULE}/Makefile
sed "s/{{.MODULE}}/${MODULE}/g" ./templates/Dockerfile > ../../${MODULE}/Dockerfile