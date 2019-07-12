#!/bin/sh
set -e

# This script is meant for build and publish deb automately
# It consist of three step:
#
#   * check build and publish development
#   * buils deb of openedge following debian's package rules
#   * publish deb package using aptly
#
# More infos, you can visit following links:
#   * https://www.debian.org/doc/manuals/maint-guide
#   * https://www.aptly.info/


# check lsb_release
if [ ! -x /usr/bin/lsb_release ]; then
    exec_cmd 'apt-get update'
    exec_cmd "apt-get install -y lsb_release > /dev/null 2>&1"
fi

# check dpkg
if [ ! -x /usr/bin/dpkg ]; then
    exec_cmd 'apt-get update'
    exec_cmd "apt-get install -y dpkg > /dev/null 2>&1"
fi

# this script is based on ubuntu16.06 amd64
if [ $(lsb_release -is) != 'Ubuntu' -o $(lsb_release -cs) != 'xenial' -o $(lsb_release -rs) != '16.04' ]; then
    echo 'Please use Ubuntu16.04 system'
    exit 1
fi

# this script is based on ubuntu16.06 amd64
if [ $(dpkg --print-architecture) != 'amd64' ]; then
    echo 'Please use amd64 arch'
    exit 1
fi

# check go installe
if $(which go); then
    echo 'Please install go'
    exit 1
fi

GOVER1=$(go version | sed -r 's/.*\bgo([0-9]+)\.([0-9]+).*\b/\1/g')
GOVER2=$(go version | sed -r 's/.*\bgo([0-9]+)\.([0-9]+).*\b/\2/g')

# check go version
if [[ ! ($GOVER1 -gt 0 && $GOVER2 -gt 10) ]]; then
    echo 'go version is too old...'
    echo 'Please use higher version!'
    exit 1
fi

PRE_INSTALL_PKGS=""

# check package gdebi-core
if $(dpkg --get-selections | grep gdebi-core); then
    PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} gdebi-core"
fi

# check package dpkg-dev
if $(dpkg --get-selections | grep dpkg-dev); then
    PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} dpkg-dev"
fi

# check package debhelper
if $(dpkg --get-selections | grep debhelper); then
    PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} debhelper"
fi

# check package dh-virtualenv
if $(dpkg --get-selections | grep dh-virtualenv); then
    PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} dh-virtualenv"
fi

# check package gnupg
# gnupg is used to sign the aptly metainfo
if $(dpkg --get-selections | grep '\bgnupg\s' -E); then
    PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} gnupg"
fi

# check package gnupg2
if $(dpkg --get-selections | grep gnupg2); then
    PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} gnupg2"
fi

# check package gnupg-agent
if $(dpkg --get-selections | grep gnupg-agent); then
    PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} gnupg-agent"
fi

# check package aptly
if $(dpkg --get-selections | grep aptly); then
    PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} aptly"
fi

# check package curl
if $(dpkg --get-selections | grep curl); then
    PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} curl"
fi

# check package curl
if $(dpkg --get-selections | grep curl); then
    PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} curl"
fi

# check package ca-certificates
if $(dpkg --get-selections | grep ca-certificates); then
    PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} ca-certificates"
fi

# check package git
if $(dpkg --get-selections | grep git); then
    PRE_INSTALL_PKGS="${PRE_INSTALL_PKGS} git"
fi

# Populating Cache
print_status "Populating apt-get cache..."
exec_cmd 'apt-get update'

if [ "X${PRE_INSTALL_PKGS}" != "X" ]; then
    print_status "Installing packages required for setup:${PRE_INSTALL_PKGS}..."
    # This next command needs to be redirected to /dev/null or the script will bork
    # in some environments
    exec_cmd "apt-get install -y${PRE_INSTALL_PKGS} > /dev/null 2>&1"
fi

# get the latest commit
go get -g github.com/baidu/openedge

cd $GOPATH/src/github.com/baidu/openedge

# check gpg-agent version
# set gpg-agent PATH
export GPG_AGENT_INFO=${HOME}/.gnupg/S.gpg-agent:0:1

# TODO: 1. gpg-agent 版本问题 2和1；2..gunpg 目录存不存在
# 而且要设置 gpg-agent 永久储存密码

# remove useless deb file
rm ../openedge_*.changes ../openedge_*.deb

# 切换到最新的迭代
# get latest tag
LatestTag=$(git describe --tags `git rev-list --tags --max-count=1`)

# check to the commit with latest tag
git checkout $LatestTag

# 提交时候要开启
# git clean -df

# TODO: generate changelog file according to CHANGELOG.md

# build deb in amd64 platform
dpkg-buildpackage -a amd64 -b -d -uc -us 

# build deb in i386 platform
dpkg-buildpackage -a i386 -b -d -uc -us

# build deb in arm64 platform
dpkg-buildpackage -a arm64 -b -d -uc -us

# build deb in armhf platform
dpkg-buildpackage -a armhf -b -d -uc -us


keylist=$(gpg --list-key)
key='cehnsheng (test) <chensheng06@baidu.com>'

# check gpg key 'openedge'
if $(echo $keylist | grep $key); then
    echo "You need a gpg key 'openedge'"
    exit 1
fi

# publish debian 
debian_dist=("buster" "jessie" "stretch" "wheezy")

REPO_LIST = $(aptly repo list)

for dist in ${debian_dist[@]};do
    if $(echo $REPO_LIST | grep openedge_debian_$dist); then
        aptly repo create -architectures amd64,arm64,i386,armhf -comment 'openedge debian $dist' -component main -distribution ${dist} openedge_debian_$dist
    fi
    debs=$(ls ../*.deb | sed 's/\.\.\///g')
    debs_exist=$(aptly repo list)
    for deb in ${debs[@]};do
        if $(echo $debs_exist | grep $deb); then
            aptly repo add openedge_debian_$dist ../$deb
        fi
    done
    if ! $(aptly publish list | grep linux/debian/$dist); then
        aptly publish drop ${dist} linux/debian
    fi
    aptly publish repo openedge_debian_$dist linux/debian
done

# create ubuntu 
ubuntu_dist=("artful" "bionic" "cosmic" "disco" "trusty" "xenial" "yakkety" "zesty")

for dist in ${ubuntu_dist[@]};do
    if $(echo $REPO_LIST | grep openedge_ubuntu_$dist); then
        aptly repo create -architectures amd64,arm64,i386,armhf -comment 'openedge ubuntu $dist' -component main -distribution $dist openedge_ubuntu_$dist
    fi
    debs=$(ls ../*.deb | sed 's/\.\.\///g')
    debs_exist=$(aptly repo list)
    for deb in ${debs[@]};do
        if $(echo $debs_exist | grep $deb); then
            aptly repo add openedge_ubuntu_$dist ../$deb
        fi
    done
    if ! $(aptly publish list | grep linux/ubuntu/$dist); then
        aptly publish drop ${dist} linux/ubuntu
    fi
    aptly publish repo openedge_ubuntu_$dist linux/ubuntu
done

# create raspbian 
raspbian_dist=("jessie" "stretch" "buster")

for dist in ${raspbian_dist[@]};do
    if $(echo $REPO_LIST | grep openedge_raspbian_$dist); then
        aptly repo create -architectures arm64,i386 -comment 'openedge raspbian $dist' -component main -distribution $dist openedge_raspbian_$dist
    fi
    debs=$(ls ../*.deb | sed 's/\.\.\///g')
    debs_exist=$(aptly repo list)
    for deb in ${debs[@]};do
        if $(echo $debs_exist | grep $deb); then
            aptly repo add openedge_raspbian_$dist ../$deb
        fi
    done
    if ! $(aptly publish list | grep linux/raspbian/$dist); then
        aptly publish drop ${dist} linux/raspbian
    fi
    aptly publish repo openedge_raspbian_$dist linux/raspbian
done
