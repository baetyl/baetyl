#!/bin/sh

set -e

PACKAGE_NAME=baetyl
EXAMPLE_PATH=http://dl.baetyl.io/example/docker

systemctl stop $PACKAGE_NAME

rm -rf /usr/local/var/db/baetyl
rm -rf /usr/local/etc/baetyl

curl -O $EXAMPLE_PATH/docker_example.tar.gz
tar -C /usr/local -xzf docker_example.tar.gz

systemctl start $PACKAGE_NAME

exit 0
