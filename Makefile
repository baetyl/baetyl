PREFIX?=/usr/local

all: openedge modules

depends:
	godep restore ./...
	find ${GOPATH}/src/github.com/docker -path '*/vendor' -type d | xargs -IX rm -r X

modules: openedge-hub openedge-function openedge-remote-mqtt

openedge:
	go build ${RACE} .

openedge-hub:
	go build ${RACE} ./module/hub/openedge-hub

openedge-function:
	go build ${RACE} ./module/function/openedge-function

openedge-remote-mqtt:
	go build ${RACE} ./module/remote/openedge-remote-mqtt

test: pubsub benchmark consistency

pubsub:
	go build ${RACE} ./tools/pubsub

benchmark:
	go build ${RACE} ./tools/benchmark

consistency:
	go build ${RACE} ./tools/consistency

install: all
	install -d -m 0755 ${PREFIX}/bin
	install -m 0755 openedge ${PREFIX}/bin/

native-install: modules install
	install -m 0755 openedge-hub ${PREFIX}/bin/
	install -m 0755 openedge-function ${PREFIX}/bin/
	install -m 0755 openedge-remote-mqtt ${PREFIX}/bin/
	install -m 0755 module/function/runtime/python2.7/openedge-function-runtime-python27 ${PREFIX}/bin
	install -m 0644 module/function/runtime/python2.7/*.py ${PREFIX}/bin
	tar cf - -C example/native . | tar xvf - -C ${PREFIX}/

uninstall:
	rm -f ${PREFIX}/bin/openedge

native-uninstall: uninstall
	rm -f ${PREFIX}/bin/openedge-hub
	rm -f ${PREFIX}/bin/openedge-function
	rm -f ${PREFIX}/bin/openedge-remote-mqtt
	rm -rf ${PREFIX}/conf

.PHONY: clean
clean:
	rm -f openedge openedge-hub openedge-function openedge-remote-mqtt
	rm -f pubsub benchmark consistency

rebuild: clean all

