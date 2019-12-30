#!/bin/bash
set -e

KEY_NAME=baetyl
PASSPHRASE=@passphrase@
PARENT_PATH=~/.aptly/public
VERSION=@version@
REVERSION=@revision@

gpg --import key.private
echo 'personal-digest-preferences SHA512'>>~/.gnupg/gpg.conf
echo 'cert-digest-algo SHA512'>>~/.gnupg/gpg.conf
echo 'default-preference-list SHA512 SHA384 SHA256 SHA224 AES256 AES192 AES CAST5 ZLIB BZIP2 ZIP Uncompressed'>>~/.gnupg/gpg.conf

repo_publish() {
    REPO_LIST=$(aptly repo list)

    # publish debian
    debian_dist=("buster" "jessie" "stretch" "wheezy")
    for dist in ${debian_dist[@]}; do
        aptly repo create -architectures amd64,arm64,i386,armhf -comment "baetyl debian $dist" -component main -distribution ${dist} baetyl_debian_$dist
        aptly repo add baetyl_debian_$dist baetyl_$VERSION-$REVERSION_*.deb
        aptly publish repo -gpg-key="$KEY_NAME" -passphrase="$PASSPHRASE" -batch baetyl_debian_$dist linux/debian
    done

    # publish ubuntu
    ubuntu_dist=("artful" "bionic" "cosmic" "disco" "trusty" "xenial" "yakkety" "zesty")
    for dist in ${ubuntu_dist[@]}; do
        aptly repo create -architectures amd64,arm64,i386,armhf -comment "baetyl ubuntu $dist" -component main -distribution ${dist} baetyl_ubuntu_$dist
        aptly repo add baetyl_ubuntu_$dist baetyl_$VERSION-$REVERSION_*.deb
        aptly publish repo -gpg-key="$KEY_NAME" -passphrase="$PASSPHRASE" -batch baetyl_ubuntu_$dist linux/ubuntu
    done

    # publish raspbian
    raspbian_dist=("buster" "jessie" "stretch")
    for dist in ${raspbian_dist[@]}; do
        aptly repo create -architectures amd64,arm64,i386,armhf -comment "baetyl raspbian $dist" -component main -distribution ${dist} baetyl_raspbian_$dist
        aptly repo add baetyl_raspbian_$dist baetyl_$VERSION-$REVERSION_*.deb
        aptly publish repo -gpg-key="$KEY_NAME" -passphrase="$PASSPHRASE" -batch baetyl_raspbian_$dist linux/raspbian
    done
}

repo_publish

mkdir -p $PARENT_PATH/linux/centos/7/x86_64/RPMS
cp scripts/centos/baetyl.repo $PARENT_PATH/linux/centos/
cp baetyl-$VERSION-$REVERSION.el7.x86_64.rpm $PARENT_PATH/linux/centos/7/x86_64/RPMS
pushd $PARENT_PATH/linux/centos/7
createrepo x86_64
popd

# mac
mkdir bin && cp baetyl bin
mkdir -p etc/baetyl && cp example/docker/etc/baetyl/conf.yml etc/baetyl && cp scripts/baetyl.plist etc/baetyl
sed -i "s/level: debug//g;" etc/baetyl/conf.yml
mkdir -p var/db/baetyl var/log/baetyl var/lib/baetyl
tar cvzf baetyl-$VERSION-darwin-amd64.tar.gz bin etc var
mkdir -p $PARENT_PATH/mac/static/x86_64
cp baetyl-$VERSION-darwin-amd64.tar.gz $PARENT_PATH/mac/static/x86_64
cp baetyl-$VERSION-darwin-amd64.tar.gz $PARENT_PATH/mac/static/x86_64/baetyl-latest-darwin-amd64.tar.gz

# example
cp scripts/baetyl.plist example/docker/etc/baetyl
tar cvzf docker_example.tar.gz -C example/docker etc var
mkdir -p $PARENT_PATH/example/$VERSION/docker
cp docker_example.tar.gz $PARENT_PATH/example/$VERSION/docker
cp -r $PARENT_PATH/example/$VERSION $PARENT_PATH/example/latest

# install
cp scripts/install.sh $PARENT_PATH
cp scripts/install_with_docker_example.sh $PARENT_PATH
cp scripts/key.public $PARENT_PATH

exit $?
