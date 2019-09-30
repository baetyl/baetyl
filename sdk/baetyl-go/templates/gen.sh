#!/bin/sh

if [[ $1 == *node* ]]
then
    sed "s/{{.MODULE}}/$1/g" ./templates/Makefile-node > ../../baetyl-$1/Makefile
elif [[ $1 == *python* ]]
then
    sed "s/{{.MODULE}}/$1/g" ./templates/Makefile-python > ../../baetyl-$1/Makefile
else
    sed "s/{{.MODULE}}/$1/g" ./templates/Makefile-go > ../../baetyl-$1/Makefile
    sed "s/{{.MODULE}}/$1/g" ./templates/Dockerfile-go > ../../baetyl-$1/Dockerfile
    sed "s/{{.MODULE}}/$1/g" ./templates/package-go.yml > ../../baetyl-$1/package.yml
fi

cat ./templates/Makefile >> ../../baetyl-$1/Makefile