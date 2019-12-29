#!/bin/bash

set -e

yum update -y && yum install -y rpmdevtools rpm-sign

mkdir -p ~/rpmbuild/RPMS ~/rpmbuild/SRPMS ~/rpmbuild/BUILD ~/rpmbuild/SOURCES ~/rpmbuild/SPECS

cp baetyl-@version@.tar.gz ~/rpmbuild/SOURCES

cp scripts/centos/baetyl.spec ~/rpmbuild/SPECS/baetyl.spec

rpmbuild -v -bb --clean ~/rpmbuild/SPECS/baetyl.spec

cp ~/rpmbuild/RPMS/x86_64/$(ls ~/rpmbuild/RPMS/x86_64/) .

