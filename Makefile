PREFIX?=/usr/local

all: openedge modules

depends:
	godep restore
	find ${GOPATH}/src/github.com/docker -path '*/vendor' -type d | xargs -IX rm -r X

depends-save:
	cd ${GOPATH}/src/github.com/docker/docker && git checkout . && cd -
	cd ${GOPATH}/src/github.com/docker/distribution && git checkout . && cd -
	godep save ./...

modules: openedge-hub openedge-function openedge-remote-mqtt

openedge:
	go build ${RACE} .

openedge-hub:
	go build ${RACE} ./modules/openedge-hub

openedge-function:
	go build ${RACE} ./modules/openedge-function

openedge-remote-mqtt:
	go build ${RACE} ./modules/openedge-remote-mqtt

test:
	go test --race ./...

tools: pubsub

pubsub:
	go build ${RACE} ./tools/pubsub

consistency:
	go build ${RACE} ./tools/consistency

install: all
	install -d -m 0755 ${PREFIX}/bin
	install -m 0755 openedge ${PREFIX}/bin/

native-install: install
	install -m 0755 openedge-hub ${PREFIX}/bin/
	install -m 0755 openedge-function ${PREFIX}/bin/
	install -m 0755 openedge-remote-mqtt ${PREFIX}/bin/
	install -m 0755 modules/openedge-function-runtime-python27/openedge_function_runtime_python27.py ${PREFIX}/bin
	install -m 0755 modules/openedge-function-runtime-python27/runtime_pb2.py ${PREFIX}/bin
	install -m 0755 modules/openedge-function-runtime-python27/runtime_pb2_grpc.py ${PREFIX}/bin
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
	rm -f openedge openedge-hub openedge-function openedge-remote-mqtt
	rm -f pubsub consistency

rebuild: clean all
