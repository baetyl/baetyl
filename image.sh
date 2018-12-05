#!/bin/sh

while [ -n "$1" ]
do
	case "$1" in
		-r)registory=$2
		   shift ;;
		-v)version=$2
           shift ;;
	esac
	shift
done

if [ "$version" = "" ];then
echo 'please input version with -v'
exit
fi
if [ "$registory" = "" ];then
echo 'please input registory with -r'
exit
fi

docker run -v ~/go/src:/go/src -w="/go/src/icode.baidu.com/baidu/bce-iot/edge-core" golang:alpine /bin/sh build.sh

for module in 'hub' 'function' 'remote_mqtt' 'function_runtime_python2.7'
# for module in 'hub' 'function'
do
name="openedge_$module"
path="module/${module//\_//}"
echo $path
image=$(docker build -f $path/Dockerfile . | grep 'Successfully built' | awk '{print $3}')
echo "$name: $image"
if [ "$image" == "" ];then
echo 'docker build failed'
exit
fi
docker tag $image $registory/$name:$version
docker tag $image $registory/$name:latest
docker push $registory/$name:$version
docker push $registory/$name:latest

done
