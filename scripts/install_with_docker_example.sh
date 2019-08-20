#!/bin/sh

set -e

PACKAGE_NAME=openedge
EXAMPLE_PATH=http://download.openedge.tech/example/docker

systemctl stop $PACKAGE_NAME

rm -rf /usr/local/var/db/openedge
rm -rf /usr/local/etc/openedge

curl -O $EXAMPLE_PATH/docker_example.tar.gz
tar -C /usr/local -xzf docker_example.tar.gz

systemctl start $PACKAGE_NAME

exit 0
