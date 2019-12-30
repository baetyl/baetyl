#!/bin/bash

set -e
EXAMPLE_PATH=http://dl.baetyl.tech/example/latest/docker
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

TARGET=$EXAMPLE_PATH/docker_example.tar.gz
if [ ${OS} = Darwin ]; then
    # docker for mac runs as a non-root user
    exec_cmd_nobail "curl $TARGET | tar xvzf - -C /usr/local/" "sudo"
    sudo chown -R $(whoami) /usr/local/etc /usr/local/var
    chmod u+x /usr/local/etc /usr/local/var
else
    exec_cmd_nobail "curl $TARGET | tar xvzf - -C /usr/local/" "sudo"
fi

print_status "Import the example configuration successfully!"

read -p "Baetyl functional modules are released as docker images. Do you want to pull required images in advance? Yes/No (default: Yes): " v
if [[ "${v}" == "n" || "${v}" == "N" || "${v}" == "no" || "${v}" == "NO" || "${v}" == "No" ]]; then
    exit 0
fi

IMAGES=$(cat /usr/local/var/db/baetyl/application.yml | grep 'image:' | grep -v grep | grep -v '#' | sed 's/.*image:\(.*\)/\1/g')
for image in ${IMAGES[@]}; do
    docker pull ${image}
done

print_status "Pull required images successfully!"

exit 0
