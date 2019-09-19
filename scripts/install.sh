#!/bin/sh

set -e

PACKAGE_NAME=baetyl
URL_PACKAGE=dl.baetyl.io
URL_KEY=http://${URL_PACKAGE}/key.public
LSB_DIST=$(. /etc/os-release && echo "$ID")
PRE_INSTALL_PKGS=""

print_status() {
    echo
    echo "## $1"
    echo
}

if [ ! -x /usr/bin/gpg ]; then
    PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} gnupg"
fi

if [ $LSB_DIST = centos ]; then

    if [ "X${PRE_INSTALL_PKGS}" != "X" ]; then
        yum install -y ${PRE_INSTALL_PKGS} >/dev/null 2>&1
    fi

    {
        echo '[baetyl]'
        echo 'name=baetyl'
        echo "baseurl=http://${URL_PACKAGE}/linux/centos/7/x86_64"
        echo 'gpgcheck=1'
        echo 'enabled=1'
        echo "gpgkey=$URL_KEY"
    } >>/etc/yum.repos.d/baetyl.repo

    yum install -y $PACKAGE_NAME
    systemctl enable $PACKAGE_NAME
else

    if [ ! -x /usr/bin/lsb_release ]; then
        PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} lsb-release"
    fi

    if [ ! -x /usr/bin/curl ]; then
        PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} curl"
    fi

    if [ ! -e /usr/lib/apt/methods/https ]; then
        PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} apt-transport-https"
    fi

    if [ "X${PRE_INSTALL_PKGS}" != "X" ]; then
        apt-get update
        apt-get install -y ${PRE_INSTALL_PKGS} >/dev/null 2>&1
    fi

    echo "deb http://${URL_PACKAGE}/linux/$(lsb_release -is | tr 'A-Z' 'a-z') $(lsb_release -cs) main" |
        tee /etc/apt/sources.list.d/${PACKAGE_NAME}.list

    curl -fsSL $URL_KEY | apt-key add -

    print_status "Added sign key!"

    apt update
    apt install $PACKAGE_NAME
fi

print_status "Install $PACKAGE_NAME Successfully!"

exit 0
