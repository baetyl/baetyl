TAG?=latest
PYSRC85=function_grpc_pb.js function_pb.js openedge-function-node85.js

package85.zip: $(PYSRC85)
	install -d -m 0755 __package/bin
	cp $^ __package/bin
	npm install
	cp -r node_modules __package/bin
	chmod +x __package/bin/openedge-function-node85.js
	cp package85.yml __package/package.yml
	cd __package && zip -q -r ../package85.zip bin/ package.yml
	rm -rf __package

image:
	@echo "IMAGE openedge-function-node85"
	docker build -t $(IMAGE_PREFIX)openedge-function-node85$(IMAGE_SUFFIX):$(TAG) -f Dockerfile85 .

.PHONY: clean
clean:
	rm -f package85.zip
	rm -rf node_modules

.PHONY: rebuild
rebuild: clean all
