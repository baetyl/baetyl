#!/bin/bash
set -e

# This script is meant for build and publish deb automately
# It consist of three step:
#
#       1. 接收参数
#       2. 导入密钥
#       3. 环境配置
#       4. 打包：最新代码；打包
#       5. 本地发布
#       6. 拷贝到远程
#
# Please run it on Ubuntu16.04 amd64 machine.
#
# More infos, you can visit following links:
#   * https://www.debian.org/doc/manuals/maint-guide
#   * https://www.aptly.info/

SCRIPT_NAME=$(basename "$0")
PACKAGE_NAME=openedge
PRIVATE_KEY_NAME=key.private

print_status() {
    echo
    echo "## $1"
}

bail() {
    echo 'Error executing command, exiting'
    exit 1
}

exec_cmd_nobail() {
    echo "+ $1"
    bash -c "$1"
}

exec_cmd() {
    exec_cmd_nobail "$1" || bail
}

usage() {
    echo "$SCRIPT_NAME [options]"
    echo "Note: you need to run this as sudo or root."
    echo ""
    echo "options"
    echo " -a, --address-repo   Address of remote publish repo, necessary"
    echo " -k, --keyname        Keyname of secret gpg key for signing, necessary"
    echo " -p, --passphrase     Passphrase for gpg-key, necessary"
    echo " -u, --url-key        Url of gpg secret key, necessary"
    echo " -w, --password       Password for current user"
    exit 1
}

print_help_and_exit() {
    print_status "Run $SCRIPT_NAME --help for more information."
    exit 1
}

process_args() {
    save_next_arg=0
    for arg in "$@"; do
        if [ $save_next_arg -eq 1 ]; then
            ADDRESS_REPO="$arg"
            save_next_arg=0
        elif [ $save_next_arg -eq 2 ]; then
            KEY_NAME="$arg"
            save_next_arg=0
        elif [ $save_next_arg -eq 3 ]; then
            PASSPHRASE="$arg"
            save_next_arg=0
        elif [ $save_next_arg -eq 4 ]; then
            URL_KEY="$arg"
            save_next_arg=0
        elif [ $save_next_arg -eq 5 ]; then
            PASSWORD="$arg"
            save_next_arg=0
        else
            case "$arg" in
            "-h" | "--help") usage ;;
            "-a" | "--address") save_next_arg=1 ;;
            "-k" | "--keyname") save_next_arg=2 ;;
            "-p" | "--passphrase") save_next_arg=3 ;;
            "-u" | "--url-key") save_next_arg=4 ;;
            "-w" | "--password") save_next_arg=5 ;;
            *) usage ;;
            esac
        fi
    done

    if [[ -z ${ADDRESS_REPO} ]]; then
        print_status "Address of remote publish repo invalid"
        print_help_and_exit
    fi

    if [[ -z ${KEY_NAME} ]]; then
        print_status "Keyname of secret gpg key for signing invalid"
        print_help_and_exit
    fi

    if [[ -z ${PASSPHRASE} ]]; then
        print_status "Passphrase for gpg-key invalid."
        print_help_and_exit
    fi

    if [[ -z ${URL_KEY} ]]; then
        print_status "Url of gpg secret key invalid."
        print_help_and_exit
    fi

    if [[ -z ${PASSWORD} ]]; then
        print_status "Password of current use invalid."
        print_help_and_exit
    fi
}

install_check_deps() {
    PRE_INSTALL_PKGS=""

    # check lsb_release
    if [ ! -x /usr/bin/lsb_release ]; then
        PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} lsb_release"
    fi

    # check dpkg
    if [ ! -x /usr/bin/dpkg ]; then
        PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} dpkg"
    fi

    # check wget
    if [ ! -x /usr/bin/wget ]; then
        PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} wget"
    fi

    # check curl
    if [ ! -x /usr/bin/curl ]; then
        PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} curl"
    fi

    if [ "X${PRE_INSTALL_PKGS}" != "X" ]; then
        exec_cmd "echo $PASSWORD | sudo -S apt-get update"

        print_status "Installing packages required for setup:${PRE_INSTALL_PKGS}..."
        # This next command needs to be redirected to /dev/null or the script will bork
        # in some environments
        exec_cmd "echo $PASSWORD | sudo -S apt-get install -y ${PRE_INSTALL_PKGS} > /dev/null 2>&1"
    fi

    print_status "Install check dependencies successfully!"
}

check_platform() {
    # this script is based on ubuntu16.06 amd64
    if [ $(lsb_release -is) != 'Ubuntu' -o $(lsb_release -cs) != 'xenial' -o $(lsb_release -rs) != '16.04' ]; then
        print_status 'Please use Ubuntu16.04 system'
        exit 1
    fi

    # this script is based on ubuntu16.06 amd64
    if [ $(dpkg --print-architecture) != 'amd64' ]; then
        print_status 'Please use amd64 arch'
        exit 1
    fi

    print_status "Machine is Ubuntu16.04 amd64"
}

check_go() {

    # check go installe
    if [[ -z $(which go) ]]; then
        exec_cmd "rm -rf go1.12.5.linux-amd64.tar.gz"

        wget https://studygolang.com/dl/golang/go1.12.5.linux-amd64.tar.gz
        exec_cmd "echo $PASSWORD | sudo -S tar -C /usr/local -xzf go1.12.5.linux-amd64.tar.gz > /dev/null 2>&1"

        # 此处必须要用单引号，双引号转移后发现不好使
        {
            echo ''
            echo 'export GOROOT=/usr/local/go'
            echo 'export GOPATH=~/go'
            echo 'export PATH=$PATH:$GOPATH:$GOROOT/bin'
        } >>$HOME/.bashrc

        # 在当前 shell 中输出变量
        export GOROOT=/usr/local/go
        export GOPATH=~/go
        export PATH=$PATH:$GOPATH:$GOROOT/bin

        print_status "Installed go1.12.5!"
    fi

    GOVER1=$(go version | sed -r 's/.*\bgo([0-9]+)\.([0-9]+).*\b/\1/g')
    GOVER2=$(go version | sed -r 's/.*\bgo([0-9]+)\.([0-9]+).*\b/\2/g')

    # check go version
    if [[ ! ($GOVER1 -gt 0 && $GOVER2 -gt 10) ]]; then
        print_status 'Go version is too old, please use higher version!'
        exit 1
    fi

    print_status "Go version: $(go version)"
}

import_key() {
    # check secret key
    if [[ ! -z $(gpg --list-secret-keys | grep $KEY_NAME) ]]; then
        print_status "Already have gpg secret key $KEY_NAME"
    else
        # 导入密钥
        curl -fsSL http://$URL_KEY/$PRIVATE_KEY_NAME | gpg --import -

        # check secret key
        if [[ -z $(gpg --list-secret-keys | grep $KEY_NAME) ]]; then
            print_status "gpg import $KEY_NAME failed"
            exit 1
        fi

        print_status "gpg import $KEY_NAME successfully"

        # 设置默认摘要算法
        # use SHA256, not SHA1
        {
            echo ''
            echo 'personal-digest-preferences SHA256'
            echo 'cert-digest-algo SHA256'
            echo 'default-preference-list SHA512 SHA384 SHA256 SHA224 AES256 AES192 AES CAST5 ZLIB BZIP2 ZIP Uncompressed'
        } >>~/.gnupg/gpg.conf
    fi
}

install_deps() {
    PRE_INSTALL_PKGS=""

    # check package gdebi-core
    if [[ -z $(dpkg --get-selections | grep gdebi-core) ]]; then
        PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} gdebi-core"
    fi

    # check package dpkg-dev
    if [[ -z $(dpkg --get-selections | grep dpkg-dev) ]]; then
        PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} dpkg-dev"
    fi

    # check package debhelper
    if [[ -z $(dpkg --get-selections | grep debhelper) ]]; then
        PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} debhelper"
    fi

    # check package dh-virtualenv
    if [[ -z $(dpkg --get-selections | grep dh-virtualenv) ]]; then
        PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} dh-virtualenv"
    fi

    # check package gnupg
    if [[ -z $(dpkg --get-selections | grep '\bgnupg\s' -E) ]]; then
        PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} gnupg"
    fi

    # check package gnupg2
    if [[ -z $(dpkg --get-selections | grep gnupg2) ]]; then
        PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} gnupg2"
    fi

    # check package aptly
    if [[ -z $(dpkg --get-selections | grep aptly) ]]; then
        PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} aptly"
    fi

    # check package ca-certificates
    if [[ -z $(dpkg --get-selections | grep ca-certificates) ]]; then
        PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} ca-certificates"
    fi

    # check package git
    if [[ -z $(dpkg --get-selections | grep git) ]]; then
        PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} git"
    fi

    if [ "X${PRE_INSTALL_PKGS}" != "X" ]; then
        exec_cmd "echo $PASSWORD | sudo -S apt-get update"

        print_status "Installing packages required for deb build: ${PRE_INSTALL_PKGS}"
        # This next command needs to be redirected to /dev/null or the script will bork
        # in some environments
        exec_cmd "echo $PASSWORD | sudo -S apt-get install -y ${PRE_INSTALL_PKGS} > /dev/null 2>&1"
    fi

    print_status "Install dependencies successfully!"
}

get_code() {
    # # get the latest commit
    # go get -v -u -x -d github.com/baidu/openedge

    # cd $GOPATH/src/github.com/baidu/openedge

    # # get latest tag
    # LatestTag=$(git describe --tags $(git rev-list --tags --max-count=1))

    # # check to the commit with latest tag
    # git checkout $LatestTag

    # # git clean -df

    # print_status "get latest release successfully!"
}

build_deb() {
    # remove useless deb file
    rm -f ../openedge_*.changes ../openedge_*.deb

    rm -rf ./debian

    cp -r ./scripts/debian ./debian

    # TODO: generate changelog file according to CHANGELOG.md

    # amd64
    sed -i "s/make install PREFIX=debian\/openedge/env GOOS=linux GOARCH=amd64 make install PREFIX=debian\/openedge/g" debian/rules
    dpkg-buildpackage -a amd64 -b -d -uc -us

    # arm64
    sed -i "s/env GOOS=linux GOARCH=amd64 make install PREFIX=debian\/openedge/env GOOS=linux GOARCH=arm64 make install PREFIX=debian\/openedge/g" debian/rules
    dpkg-buildpackage -a arm64 -b -d -uc -us

    # i386
    sed -i "s/env GOOS=linux GOARCH=arm64 make install PREFIX=debian\/openedge/env GOOS=linux GOARCH=386 make install PREFIX=debian\/openedge/g" debian/rules
    dpkg-buildpackage -a i386 -b -d -uc -us

    # armhf
    sed -i "s/env GOOS=linux GOARCH=386 make install PREFIX=debian\/openedge/env GOOS=linux GOARCH=arm GOARM=7 make install PREFIX=debian\/openedge/g" debian/rules
    dpkg-buildpackage -a armhf -b -d -uc -us

    print_status "build_deb successfully!"
}

repo_publish() {

    REPO_LIST=$(aptly repo list)

    # publish debian
    debian_dist=("buster" "jessie" "stretch" "wheezy")

    for dist in ${debian_dist[@]}; do
        if [[ -z $(echo $REPO_LIST | grep openedge_debian_$dist) ]]; then
            aptly repo create -architectures amd64,arm64,i386,armhf -comment "openedge debian $dist" -component main -distribution ${dist} openedge_debian_$dist
        fi
        debs=$(ls ../*.deb | sed 's/\.\.\///g; s/\.deb//g;')
        debs_exist=$(aptly repo show -with-packages openedge_debian_$dist)
        for deb in ${debs[@]}; do
            if [[ -z $(echo $debs_exist | grep $deb) ]]; then
                aptly repo add openedge_debian_$dist ../$deb
            fi
        done
        if [[ ! -z $(aptly publish list | grep linux/debian/$dist) ]]; then
            aptly publish drop ${dist} linux/debian
        fi
        aptly publish repo -gpg-key="$GPG_KEY" -passphrase="$PASSPHRASE" openedge_debian_$dist linux/debian
    done

    # publish ubuntu
    ubuntu_dist=("artful" "bionic" "cosmic" "disco" "trusty" "xenial" "yakkety" "zesty")

    for dist in ${ubuntu_dist[@]}; do
        if [[ -z $(echo $REPO_LIST | grep openedge_ubuntu_$dist) ]]; then
            aptly repo create -architectures amd64,arm64,i386,armhf -comment "openedge ubuntu $dist" -component main -distribution ${dist} openedge_ubuntu_$dist
        fi
        debs=$(ls ../*.deb | sed 's/\.\.\///g; s/\.deb//g;')
        debs_exist=$(aptly repo show -with-packages openedge_ubuntu_$dist)
        for deb in ${debs[@]}; do
            if [[ -z $(echo $debs_exist | grep $deb) ]]; then
                aptly repo add openedge_ubuntu_$dist ../$deb
            fi
        done
        if [[ ! -z $(aptly publish list | grep linux/ubuntu/$dist) ]]; then
            aptly publish drop ${dist} linux/ubuntu
        fi
        aptly publish repo -gpg-key="$GPG_KEY" -passphrase="$PASSPHRASE" openedge_ubuntu_$dist linux/ubuntu
    done

    # create raspbian
    raspbian_dist=("jessie" "stretch" "buster")

    for dist in ${raspbian_dist[@]}; do
        if ! $(echo $REPO_LIST | grep openedge_raspbian_$dist); then
            aptly repo create -architectures arm64,i386 -comment "openedge raspbian $dist" -component main -distribution $dist openedge_raspbian_$dist
        fi
        debs=$(ls ../*.deb | sed 's/\.\.\///g')
        debs_exist=$(aptly repo list)
        for deb in ${debs[@]}; do
            if $(echo $debs_exist | grep $deb); then
                aptly repo add openedge_raspbian_$dist ../$deb
            fi
        done
        if $(aptly publish list | grep linux/raspbian/$dist); then
            aptly publish drop ${dist} linux/raspbian
        fi
        aptly publish repo -gpg-key="$GPG_KEY" -passphrase="$PASSPHRASE" openedge_raspbian_$dist linux/raspbian
    done

    # publish raspbian
    raspbian_dist=("artful" "bionic" "cosmic" "disco" "trusty" "xenial" "yakkety" "zesty")

    for dist in ${raspbian_dist[@]}; do
        if [[ -z $(echo $REPO_LIST | grep openedge_raspbian_$dist) ]]; then
            aptly repo create -architectures amd64,arm64,i386,armhf -comment "openedge raspbian $dist" -component main -distribution ${dist} openedge_raspbian_$dist
        fi
        debs=$(ls ../*.deb | sed 's/\.\.\///g; s/\.deb//g;')
        debs_exist=$(aptly repo show -with-packages openedge_raspbian_$dist)
        for deb in ${debs[@]}; do
            if [[ -z $(echo $debs_exist | grep $deb) ]]; then
                aptly repo add openedge_raspbian_$dist ../$deb
            fi
        done
        if [[ ! -z $(aptly publish list | grep linux/raspbian/$dist) ]]; then
            aptly publish drop ${dist} linux/raspbian
        fi
        aptly publish repo -gpg-key="$GPG_KEY" -passphrase="$PASSPHRASE" openedge_raspbian_$dist linux/raspbian
    done
}

# #拷贝到 远程官方机器
# scp xx xx@xx:xx
# # 单纯拷贝

process_args "$@"

install_check_deps

check_platform

check_go

import_key

install_deps

get_code

build_deb

repo_publish

exit $?
