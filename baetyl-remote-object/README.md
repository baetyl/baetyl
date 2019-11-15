# baetyl-remote-object

Baetyl-remote-object module which is used to upload object(maybe image, file, video etc.) to some services like [BOS(Baidu Object Storage)](https://cloud.baidu.com/product/bos.html), [COS(Ceph Object Storage)](https://ceph.io/) and [Amazon S3(Simple Storage Service)](https://aws.amazon.com/s3/).

## Build

```shell
cd ../.. # change directory to the path of baetyl
make rebuild MODULES="remote-object" # build baetyl-remote-objcet
make rebuild PLATFORMS=all MODULES="remote-object" # build baetyl-remote-object of all supported platforms
make rebuild PLATFORMS="linux/amd64" MODULES="remote-object" # build baetyl-remote-object of specified platform

make image MODULES="remote-object" # build docker image for baetyl-remote-object module
make image PLATFORMS=all MODULES="remote-object" # build docker images of all supported platforms for baetyl-remote-object module
make image PLATFORMS="linux/amd64" MODULES="remote-object" # build docker image of specified platform for baetyl-remote-object module
```

## Example

More detailed contents of configuration please refer to example folder.
