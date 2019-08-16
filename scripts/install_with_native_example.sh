#!/bin/sh

set -e

PACKAGE_NAME=openedge
EXAMPLE_PATH=http://download.openedge.tech/example/native

systemctl stop $PACKAGE_NAME

rm -rf /usr/local/var/db/openedge
rm -rf /usr/local/etc/openedge

curl -O $EXAMPLE_PATH/native_example.tar.gz
tar -C /usr/local -xzf native_example.tar.gz

systemctl start $PACKAGE_NAME

exit 0
