#!/bin/bash

set -e

workdir=$(
    cd $(dirname $0)
    pwd
)
cd $workdir

PACKAGE_NAME=baetyl
VERSION=0.1.6
EXAMPLE_PATH=http://dl.baetyl.io/example/${VERSION}/docker
OS=$(uname)

print_status() {
    echo
    echo "## $1"
    echo
}

exec_cmd_nobail() {
    echo "+ $2 bash -c '$1'"
    $2 bash -c "$1"
}

if ls /usr/local/var/db/baetyl >/dev/null 2>&1; then
    read -p "There are outdated configurations. Do you want to remove or not (y/n, default: y): " v
    if [[ "${v}" == "n" || "${v}" == "N" || "${v}" == "no" || "${v}" == "NO" || "${v}" == "No" ]]; then
        print_status "This script will exit now..."
        exit 1
    fi
    exec_cmd_nobail "rm -rf /usr/local/etc/baetyl /usr/local/var/db/baetyl" "sudo"
fi

if [ ${OS} = Darwin ]; then
    # docker for mac runs as a non-root user
    TEMPDIR=$(mktemp -d)
    exec_cmd_nobail "curl $EXAMPLE_PATH/docker_example.tar.gz | tar xvzf - -C ${TEMPDIR}"
    exec_cmd_nobail "tar cf - -C ${TEMPDIR} etc var | tar xvf - -C /usr/local/" "sudo"
    rm -rf ${TEMPDIR}
else
    exec_cmd_nobail "curl $EXAMPLE_PATH/docker_example.tar.gz | tar xvzf - -C /usr/local/" "sudo"
fi

print_status "Import the example configuration Successfully!"

read -p "Baetyl functional modules are released as docker images. Do you want to pull required images now (y/n, default: y): " v
if [[ "${v}" == "n" || "${v}" == "N" || "${v}" == "no" || "${v}" == "NO" || "${v}" == "No" ]]; then
    exit 0
fi

TARGETS=(hub function-manager function-python27 function-python36 function-node85 function-sql timer)
for target in ${TARGETS[@]}; do
    docker pull hub.baidubce.com/baetyl/baetyl-${target}
done

print_status "Pulled required images Successfully!"

exit 0
