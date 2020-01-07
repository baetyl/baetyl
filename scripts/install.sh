#!/bin/bash

set -e

NAME=baetyl
URL_PACKAGE=$1
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
    echo "+ $2 bash -c '$1'"
    $2 bash -c "$1"
}

if [ -x "$(command -v gpg)" ]; then
    PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} gnupg"
fi

if [ ${OS} = Darwin ]; then
    TARGET=http://${URL_PACKAGE}/mac/static/x86_64/${NAME}-latest-darwin-amd64.tar.gz
    exec_cmd_nobail "curl $TARGET | tar xvzf - -C /usr/local" "sudo"
    sudo chown -R $(whoami) /usr/local/bin /usr/local/etc /usr/local/var
    chmod u+x /usr/local/bin /usr/local/etc /usr/local/var
    chmod +x /usr/local/bin/baetyl
else
    LSB_DIST=$(. /etc/os-release && echo "$ID" | tr '[:upper:]' '[:lower:]')

    case "$LSB_DIST" in
    ubuntu | debian | raspbian)
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
            exec_cmd_nobail "apt-get update" "sudo"
            exec_cmd_nobail "apt-get install -y ${PRE_INSTALL_PKGS} >/dev/null 2>&1" "sudo"
        fi

        exec_cmd_nobail "echo \"deb http://${URL_PACKAGE}/linux/${LSB_DIST} $(lsb_release -cs) main\" |
        tee /etc/apt/sources.list.d/${NAME}.list" "sudo"

        exec_cmd_nobail "curl -fsSL ${URL_KEY} | apt-key add -" "sudo"

        print_status "Added sign key!"

        exec_cmd_nobail "apt update" "sudo"
        exec_cmd_nobail "apt install ${NAME}" "sudo"
        ;;
    centos)
        PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} yum-utils"
        YUM_REPO="http://${URL_PACKAGE}/linux/$LSB_DIST/${NAME}.repo"

        if [ "X${PRE_INSTALL_PKGS}" != "X" ]; then
            exec_cmd_nobail "yum install -y ${PRE_INSTALL_PKGS}" "sudo"
        fi

        if ! curl -Ifs "$YUM_REPO" >/dev/null; then
            print_status "Error: Unable to curl repository file $YUM_REPO, is it valid?"
            exit 1
        fi

        exec_cmd_nobail "yum-config-manager --add-repo $YUM_REPO" "sudo"
        exec_cmd_nobail "yum makecache" "sudo"

        exec_cmd_nobail "yum install -y ${NAME}" "sudo"
        exec_cmd_nobail "systemctl enable ${NAME}" "sudo"
        ;;
    *)
        print_status "Your OS is not supported!"
        ;;
    esac
fi

print_status "Install ${NAME} successfully!"

exit 0
