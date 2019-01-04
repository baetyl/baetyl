PREFIX?=/usr/local

all: openedge-function-runtime-node

openedge-function-runtime-node:
	@echo "build $@"
	npm install

openedge-function-runtime-node-image:
	@echo "build ${GOFLAG} $@"
	docker build -t openedge-function-runtime-node85:build -f Dockerfile85 .
	docker build -t openedge-function-runtime-node611:build -f Dockerfile611 .
	docker build -t openedge-function-runtime-node1015:build -f Dockerfile1015 .

.PHONY: clean
clean:
	rm -rf node_modules
	rm -rf package-lock.json

.PHONY: rebuild
rebuild: clean all

install: all
	install -d -m 0755 ${PREFIX}/bin
	install -m 0755 openedge_function_runtime_grpc_pb.js ${PREFIX}/bin
	install -m 0755 openedge_function_runtime_node.js ${PREFIX}/bin
	install -m 0755 openedge_function_runtime_pb.js ${PREFIX}/bin
	tar cf - node_modules | tar xvf - -C ${PREFIX}/bin

uninstall:
	rm -f ${PREFIX}/bin/openedge_function_runtime_pb.js
	rm -f ${PREFIX}/bin/openedge_function_runtime_grpc_pb.js
	rm -f ${PREFIX}/bin/openedge_function_runtime_node.js
	rm -rf ${PREFIX}/bin/node_modules