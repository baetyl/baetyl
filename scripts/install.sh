#!/bin/bash

set -e

workdir=$(
    cd $(dirname $0)
    pwd
)
cd $workdir

PACKAGE_NAME=baetyl
VERSION=0.1.6
URL_PACKAGE=dl.baetyl.io
URL_KEY=http://${URL_PACKAGE}/key.public
OS=$(uname)
PRE_INSTALL_PKGS="ca-certificates"
EFFECTIVE_UID=$("id" | grep -o "uid=[0-9]*" | cut -d= -f2)

print_status() {
    echo
    echo "## $1"
    echo
}

exec_cmd_nobail() {
    echo "+ bash -c '$1'"
    bash -c "$1"
}

if [ $EFFECTIVE_UID -ne 0 ]; then
    print_status "The script needs to be run as root."
    exit 1
fi

if [ -x "$(command -v gpg)" ]; then
    PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} gnupg"
fi

if [ ${OS} = Darwin ]; then
    exec_cmd_nobail "curl http://${URL_PACKAGE}/mac/static/x86_64/baetyl-${VERSION}-darwin-amd64.tar.gz | tar xvzf - -C /usr/local"
else
    LSB_DIST=$(. /etc/os-release && echo "$ID")

    if [ ${LSB_DIST} = centos ]; then
        PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} yum-utils"
        YUM_REPO="http://${URL_PACKAGE}/linux/$LSB_DIST/baetyl.repo"

        if [ "X${PRE_INSTALL_PKGS}" != "X" ]; then
            exec_cmd_nobail "yum install -y ${PRE_INSTALL_PKGS}"
        fi

        if ! curl -Ifs "$YUM_REPO" >/dev/null; then
            print_status "Error: Unable to curl repository file $YUM_REPO, is it valid?"
            exit 1
        fi

        exec_cmd_nobail "yum-config-manager --add-repo $YUM_REPO"
        exec_cmd_nobail "yum makecache"

        exec_cmd_nobail "yum install -y ${PACKAGE_NAME}"
        exec_cmd_nobail "systemctl enable ${PACKAGE_NAME}"
    else
        if [ -x "$(command -v lsb_release)" ]; then
            PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} lsb-release"
        fi

        if [ -x "$(command -v curl)" ]; then
            PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} curl"
        fi

        if [ ! -e /usr/lib/apt/methods/https ]; then
            PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} apt-transport-https"
        fi

        if [ "X${PRE_INSTALL_PKGS}" != "X" ]; then
            exec_cmd_nobail "apt-get update"
            exec_cmd_nobail "apt-get install -y ${PRE_INSTALL_PKGS} >/dev/null 2>&1"
        fi

        exec_cmd_nobail "echo \"deb http://${URL_PACKAGE}/linux/${LSB_DIST} $(lsb_release -cs) main\" |
        tee /etc/apt/sources.list.d/${PACKAGE_NAME}.list"

        exec_cmd_nobail "curl -fsSL ${URL_KEY} | apt-key add -"

        print_status "Added sign key!"

        exec_cmd_nobail "apt update"
        exec_cmd_nobail "apt install ${PACKAGE_NAME}"
    fi
fi

print_status "Install ${PACKAGE_NAME} Successfully!"

exit 0
