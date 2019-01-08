all: openedge-function-runtime-node

openedge-function-runtime-node:
	@echo "build $@"
	cd openedge-function-runtime-node-lib && npm install

openedge-function-runtime-node-image:
	@echo "build $@"
	docker build -t openedge-function-runtime-node85:build -f Dockerfile85 .
	docker build -t openedge-function-runtime-node611:build -f Dockerfile611 .
	docker build -t openedge-function-runtime-node1015:build -f Dockerfile1015 .

.PHONY: clean
clean:
	rm -rf openedge-function-runtime-node-lib/node_modules
	rm -rf openedge-function-runtime-node-lib/package-lock.json

.PHONY: rebuild
rebuild: clean all