PREFIX?=/usr/local

all: openedge modules

depends:
	godep restore
	find ${GOPATH}/src/github.com/docker -path '*/vendor' -type d | xargs -IX rm -r X

depends-save:
	cd ${GOPATH}/src/github.com/docker/docker && git checkout . && cd -
	cd ${GOPATH}/src/github.com/docker/distribution && git checkout . && cd -
	godep save ./...

modules: openedge-hub/openedge-hub openedge-function/openedge-function openedge-remote-mqtt/openedge-remote-mqtt

openedge:
	@echo "GO $@ ${GOFLAG}"
	@go build ${GOFLAG} .

openedge-hub/openedge-hub:
	make -C openedge-hub

openedge-function/openedge-function:
	make -C openedge-function

openedge-remote-mqtt/openedge-remote-mqtt:
	make -C openedge-remote-mqtt

test:
	go test --race ./...

tools: pubsub openedge-consistency

pubsub:
	@echo "GO $@"
	@go build ${GOFLAG} ./tools/pubsub

openedge-consistency:
	@echo "GO $@"
	@go build ${GOFLAG} ./tools/openedge-consistency

install: all
	install -d -m 0755 ${PREFIX}/bin
	install -m 0755 openedge ${PREFIX}/bin/

native-install: install
	install -m 0755 openedge-hub/openedge-hub ${PREFIX}/bin/
	install -m 0755 openedge-function/openedge-function ${PREFIX}/bin/
	install -m 0755 openedge-remote-mqtt/openedge-remote-mqtt ${PREFIX}/bin/
	install -m 0755 openedge-function-runtime-python27/openedge_function_runtime_python27.py ${PREFIX}/bin
	install -m 0755 openedge-function-runtime-python27/runtime_pb2.py ${PREFIX}/bin
	install -m 0755 openedge-function-runtime-python27/runtime_pb2_grpc.py ${PREFIX}/bin
	tar cf - -C example/native app conf | tar xvf - -C ${PREFIX}/

uninstall:
	rm -f ${PREFIX}/bin/openedge

native-uninstall: uninstall
	rm -f ${PREFIX}/bin/openedge-hub
	rm -f ${PREFIX}/bin/openedge-function
	rm -f ${PREFIX}/bin/openedge-remote-mqtt
	rm -f ${PREFIX}/bin/runtime_pb2.py
	rm -f ${PREFIX}/bin/runtime_pb2.pyc
	rm -f ${PREFIX}/bin/runtime_pb2_grpc.py
	rm -f ${PREFIX}/bin/runtime_pb2_grpc.pyc
	rm -f ${PREFIX}/bin/openedge_function_runtime_python27.py
	rm -f ${PREFIX}/bin/openedge_function_runtime_python27.pyc
	rm -rf ${PREFIX}/conf
	rm -rf ${PREFIX}/app
	rmdir ${PREFIX}/bin
	rmdir ${PREFIX}

.PHONY: clean
clean:
	rm -f openedge
	make -C openedge-hub clean
	make -C openedge-function clean
	make -C openedge-remote-mqtt clean
	rm -f pubsub openedge-consistency

rebuild: clean all
