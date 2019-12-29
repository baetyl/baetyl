#!/bin/bash
set -e

PACKAGE_NAME=baetyl
KEY_NAME=baetyl
PASSPHRASE=oortjksdop
PARENT_PATH=/root/.aptly/public
VERSION=1.0.0
REVERSION=1

repo_publish() {

    REPO_LIST=$(aptly repo list)

    # publish debian
    debian_dist=("buster" "jessie" "stretch" "wheezy")

    for dist in ${debian_dist[@]}; do
        if [[ -z $(echo $REPO_LIST | grep baetyl_debian_$dist) ]]; then
            aptly repo create -architectures amd64,arm64,i386,armhf -comment "baetyl debian $dist" -component main -distribution ${dist} baetyl_debian_$dist
        fi
        aptly repo add baetyl_debian_$dist /package/baetyl_$VERSION-$REVERSION_*.deb
        if [[ -z $(aptly publish list | grep linux/debian/$dist) ]]; then
            aptly publish repo -gpg-key="$KEY_NAME" -passphrase="$PASSPHRASE" baetyl_debian_$dist linux/debian
        else
            aptly publish update -gpg-key="$KEY_NAME" -passphrase="$PASSPHRASE" ${dist} linux/debian
        fi
    done

    # publish ubuntu
    ubuntu_dist=("artful" "bionic" "cosmic" "disco" "trusty" "xenial" "yakkety" "zesty")

    for dist in ${ubuntu_dist[@]}; do
        if [[ -z $(echo $REPO_LIST | grep baetyl_ubuntu_$dist) ]]; then
            aptly repo create -architectures amd64,arm64,i386,armhf -comment "baetyl ubuntu $dist" -component main -distribution ${dist} baetyl_ubuntu_$dist
        fi
        aptly repo add baetyl_ubuntu_$dist ./package/baetyl_$VERSION-$REVERSION_*.deb
        if [[ -z $(aptly publish list | grep linux/ubuntu/$dist) ]]; then
            aptly publish repo -gpg-key="$KEY_NAME" -passphrase="$PASSPHRASE" baetyl_ubuntu_$dist linux/ubuntu
        else
            aptly publish update -gpg-key="$KEY_NAME" -passphrase="$PASSPHRASE" ${dist} linux/ubuntu
        fi
    done

    # publish raspbian
    raspbian_dist=("buster" "jessie" "stretch")

    for dist in ${raspbian_dist[@]}; do
        if [[ -z $(echo $REPO_LIST | grep baetyl_raspbian_$dist) ]]; then
            aptly repo create -architectures amd64,arm64,i386,armhf -comment "baetyl raspbian $dist" -component main -distribution ${dist} baetyl_raspbian_$dist
        fi
        aptly repo add baetyl_raspbian_$dist ./package/baetyl_$VERSION-$REVERSION_*.deb
        if [[ -z $(aptly publish list | grep linux/raspbian/$dist) ]]; then
            aptly publish repo -gpg-key="$KEY_NAME" -passphrase="$PASSPHRASE" baetyl_raspbian_$dist linux/raspbian
        else
            aptly publish update -gpg-key="$KEY_NAME" -passphrase="$PASSPHRASE" ${dist} linux/raspbian
        fi
    done
}

repo_publish

mkdir -p $PARENT_PATH/linux/centos/7/x86_64/RPMS
cp /package/baetyl-$VERSION-$REVERSION.el7.x86_64.rpm $PARENT_PATH/linux/centos/7/x86_64/RPMS

cd $PARENT_PATH/linux/centos/7
createrepo x86_64

# mac zip
mkdir -p $PARENT_PATH/mac/static/x86_64 && cd $PARENT_PATH/mac/static/x86_64
mkdir bin etc && mkdir etc/baetyl
cp /baetyl/output/darwin/amd64/baetyl/bin/baetyl bin
cp /baetyl/example/docker/etc/baetyl/conf.yml etc/baetyl
sed -i "s/level: debug//g;" etc/baetyl/conf.yml
cp /baetyl/scripts/baetyl.plist etc/baetyl
tar cvzf baetyl-$VERSION-darwin-amd64.tar.gz bin etc
rm -rf bin etc
ln -s baetyl-$VERSION-darwin-amd64.tar.gz baetyl-latest-darwin-amd64.tar.gz

# example zip
mkdir -p $PARENT_PATH/example/$VERSION/docker && cd $PARENT_PATH/example/$VERSION/docker
cp -r /baetyl/example/docker/* .
cp /baetyl/scripts/baetyl.plist etc/baetyl
mkdir -p var/log/baetyl
tar cvzf docker_example.tar.gz etc var
rm -rf etc var
cd $PARENT_PATH/example
ln -s $VERSION latest

cp /baetyl/scripts/install.sh $PARENT_PATH
cp /baetyl/scripts/install_with_docker_example.sh $PARENT_PATH

exit 0
