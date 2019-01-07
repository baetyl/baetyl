PREFIX?=/usr/local

all: openedge-function-runtime-node

openedge-function-runtime-node:
	@echo "build $@"
	cd openedge-function-runtime-node-lib && npm install

openedge-function-runtime-node-image:
	@echo "build ${GOFLAG} $@"
	docker build -t openedge-function-runtime-node85:build -f Dockerfile85 .
	docker build -t openedge-function-runtime-node611:build -f Dockerfile611 .
	docker build -t openedge-function-runtime-node1015:build -f Dockerfile1015 .

.PHONY: clean
clean:
	rm -rf openedge-function-runtime-node-lib/node_modules
	rm -rf openedge-function-runtime-node-lib/package-lock.json

.PHONY: rebuild
rebuild: clean all

install: all
	install -d -m 0755 ${PREFIX}/bin
	install -m 0755 openedge_function_runtime_node.js ${PREFIX}/bin
	tar cf - openedge-function-runtime-node-lib | tar xvf - -C ${PREFIX}/bin

uninstall:
	rm -f ${PREFIX}/bin/openedge_function_runtime_node.js
	rm -rf ${PREFIX}/bin/openedge-function-runtime-node-lib