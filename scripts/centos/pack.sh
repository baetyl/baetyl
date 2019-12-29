#!/bin/bash

set -e

yum update -y && yum install -y rpmdevtools rpm-sign

gpg --import private.key && rpm --import public.key

mkdir -p ~/rpmbuild/RPMS ~/rpmbuild/SRPMS ~/rpmbuild/BUILD ~/rpmbuild/SOURCES ~/rpmbuild/SPECS

cp baetyl-@version@.tar.gz ~/rpmbuild/SOURCES

cp scripts/centos/baetyl.spec ~/rpmbuild/SPECS/baetyl.spec

rpmbuild -v -bb --clean ~/rpmbuild/SPECS/baetyl.spec

echo "yes" | setsid rpm \
    --define "_gpg_name baetyl" \
    --define "_signature gpg" \
    --define "__gpg_check_password_cmd /bin/true" \
    --define "__gpg_sign_cmd %{__gpg} gpg --batch --no-armor --digest-algo 'sha512' --passphrase '@passphrase@' --no-secmem-warning -u '%{_gpg_name}' --sign --detach-sign --output %{__signature_filename} %{__plaintext_filename}" \
    --resign "$(echo ~)/rpmbuild/RPMS/x86_64/$(ls ~/rpmbuild/RPMS/x86_64)"

cp ~/rpmbuild/RPMS/x86_64/$(ls ~/rpmbuild/RPMS/x86_64/) .

rpm --checksig ~/rpmbuild/RPMS/x86_64/$(ls ~/rpmbuild/RPMS/x86_64/)

curl -O https://baetyl-repo-pre.gz.bcebos.com/linux/centos/7/x86_64/RPMS/baetyl-0.1.6-1.el7.x86_64.rpm

rpm --checksig baetyl-0.1.6-1.el7.x86_64.rpm

exit 0

