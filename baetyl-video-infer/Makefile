PREFIX?=/usr/local
TAG?=latest
SRC=$(wildcard *.go)

all: baetyl-video-infer package.zip

baetyl-video-infer: $(SRC)
	echo "BUILD $@"
	go build ${GOFLAG} .

package.zip: baetyl-video-infer
	install -d -m 0755 __package/bin
	cp $< __package/bin
	cp package.yml __package/
	cd __package && zip -q -r ../package.zip bin/ package.yml
	rm -rf __package

image: baetyl-video-infer
	echo "IMAGE $<"
	docker build -t $(IMAGE_PREFIX)$<$(IMAGE_SUFFIX):$(TAG) .

.PHONY: clean
clean:
	rm -f baetyl-video-infer package.zip

.PHONY: rebuild
rebuild: clean all

install:
	tar cf - -C example/docker etc var | tar xvf - -C ${PREFIX}/
	curl -Lo ${PREFIX}/var/db/baetyl/video-infer-model/frozen_inference_graph.pb https://baetyl.cdn.bcebos.com/ai/video-infer-model/frozen_inference_graph.pb
	curl -Lo ${PREFIX}/var/db/baetyl/video-infer-model/ssd_mobilenet_v1_coco_2017_11_17.pbtxt https://baetyl.cdn.bcebos.com/ai/video-infer-model/ssd_mobilenet_v1_coco_2017_11_17.pbtxt


install-native: package.zip
	install -d -m 0755 ${PREFIX}/var/db/baetyl/baetyl-video-infer
	unzip -o package.zip -d ${PREFIX}/var/db/baetyl/baetyl-video-infer
	tar cf - -C example/native etc var | tar xvf - -C ${PREFIX}/
	curl -Lo ${PREFIX}/var/db/baetyl/video-infer-model/frozen_inference_graph.pb https://baetyl.cdn.bcebos.com/ai/video-infer-model/frozen_inference_graph.pb
	curl -Lo ${PREFIX}/var/db/baetyl/video-infer-model/ssd_mobilenet_v1_coco_2017_11_17.pbtxt https://baetyl.cdn.bcebos.com/ai/video-infer-model/ssd_mobilenet_v1_coco_2017_11_17.pbtxt


uninstall-native:
	rm -rf ${PREFIX}/var/db/baetyl/baetyl-video-infer
