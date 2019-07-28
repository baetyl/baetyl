#!/bin/bash
set -e

print_status() {
    echo
    echo "## $1"
    echo
}

PRE_INSTALL_PKGS=""
SYSTEM_NAME=$(lsb_release -is | tr 'A-Z' 'a-z')
DISTRO=$(lsb_release -cs)
PACKAGE_NAME=openedge
URL_KEY=https://github.com/chensheng0/testfork/releases/download/key
PRUBLIC_KEY_NAME=key.public

if [ ! -x /usr/bin/lsb_release ]; then
    PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} lsb-release"
fi

if [ ! -x /usr/bin/curl ]; then
    PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} curl"
fi

if [ ! -x /usr/bin/gpg ]; then
    PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} gnupg"
fi

if [ $SYSTEM_NAME = centos ]; then

    if [ "X${PRE_INSTALL_PKGS}" != "X" ]; then
        print_status "Installing packages required for setup:${PRE_INSTALL_PKGS}..."
        # This next command needs to be redirected to /dev/null or the script will bork
        # in some environments
        yum install -y ${PRE_INSTALL_PKGS} >/dev/null 2>&1
    fi

    {
        echo '[openedge]'
        echo 'name=openedge'
        echo 'baseurl=http://106.13.23.136/linux/centos/7/x86_64'
        echo 'gpgcheck=1'
        echo 'enabled=1'
        echo 'gpgkey=https://github.com/chensheng0/testfork/releases/download/key/key.public'
    } >>/etc/yum.repos.d/openedge.repo

    yum install $PACKAGE_NAME
else

    if [ ! -e /usr/lib/apt/methods/https ]; then
        PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} apt-transport-https"
    fi

    # Populating Cache
    apt-get update

    if [ "X${PRE_INSTALL_PKGS}" != "X" ]; then
        print_status "Installing packages required for setup:${PRE_INSTALL_PKGS}..."
        # This next command needs to be redirected to /dev/null or the script will bork
        # in some environments
        apt-get install -y ${PRE_INSTALL_PKGS} >/dev/null 2>&1
    fi

    echo "deb http://106.13.23.136/linux/${SYSTEM_NAME} $DISTRO main" |
        sudo tee /etc/apt/sources.list.d/${PACKAGE_NAME}.list

    print_status "Added repo /etc/apt/sources.list.d/${PACKAGE_NAME}.list"

    curl -fsSL $URL_KEY/$PRUBLIC_KEY_NAME | sudo apt-key add -

    print_status "Added sign key"

    # verify gpg key is correct

    apt update
    apt install $PACKAGE_NAME
fi

print_status "Installed $PACKAGE_NAME!"

exit 1
