#!/bin/bash
set -e

print_status() {
    echo
    echo "## $1"
    echo
}

PRE_INSTALL_PKGS=""

# Check that HTTPS transport is available to APT
# (Check snaked from: https://get.docker.io/ubuntu/)

if [ ! -e /usr/lib/apt/methods/https ]; then
    PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} apt-transport-https"
fi

if [ ! -x /usr/bin/lsb_release ]; then
    PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} lsb-release"
fi

if [ ! -x /usr/bin/curl ]; then
    PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} curl"
fi

# Used by apt-key to add new keys

if [ ! -x /usr/bin/gpg ]; then
    PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} gnupg"
fi

# Populating Cache
apt-get update

if [ "X${PRE_INSTALL_PKGS}" != "X" ]; then
    print_status "Installing packages required for setup:${PRE_INSTALL_PKGS}..."
    # This next command needs to be redirected to /dev/null or the script will bork
    # in some environments
    apt-get install -y ${PRE_INSTALL_PKGS} >/dev/null 2>&1
fi

SYSTEM_NAME=$(lsb_release -is | tr 'A-Z' 'a-z')
DISTRO=$(lsb_release -cs)
PACKAGE_NAME=openedge
URL_KEY=https://github.com/chensheng0/testfork/releases/download/key
PRUBLIC_KEY_NAME=key.public

echo "deb http://106.13.24.234/linux/${SYSTEM_NAME} $DISTRO main" |
    sudo tee /etc/apt/sources.list.d/${PACKAGE_NAME}.list

print_status "Added repo /etc/apt/sources.list.d/${PACKAGE_NAME}.list"

curl -fsSL $URL_KEY/$PRUBLIC_KEY_NAME | sudo apt-key add -

print_status "Added sign key"

# verify gpg key is correct

apt update
apt install $PACKAGE_NAME

print_status "Installed $PACKAGE_NAME!"

exit 1

/key.public
