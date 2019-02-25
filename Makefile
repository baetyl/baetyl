PREFIX?=/usr/local
VERSION?=git-$(shell git rev-list HEAD|head -1|cut -c 1-6)
PACKAGE_PREFIX?=
GOFLAG?=-ldflags "-X main.BuildTime=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X 'main.GoVersion=`go version`' -X 'main.Version=$(VERSION)' -X 'master.Version=$(VERSION)'"

all: openedge

package: \
	openedge-agent/package.tar.gz \
	openedge-hub/package.tar.gz \
	openedge-function/package.tar.gz \
	openedge-function-python27/package.tar.gz \
	openedge-remote-mqtt/package.tar.gz

SRC=$(wildcard *.go) $(shell find master sdk-go protocol utils -type f -name '*.go')

openedge: $(SRC)
	@echo "BUILD $@"
	@go build ${GOFLAG} .

openedge-agent/package.tar.gz:
	make -C openedge-agent

openedge-hub/package.tar.gz:
	make -C openedge-hub

openedge-function/package.tar.gz:
	make -C openedge-function

openedge-function-python27/package.tar.gz:
	make -C openedge-function-python27

openedge-remote-mqtt/package.tar.gz:
	make -C openedge-remote-mqtt

test:
	go test --race ./...

tools: pubsub openedge-consistency

pubsub:
	@echo "BUILD $@"
	@go build ${GOFLAG} ./tools/pubsub

openedge-consistency:
	@echo "BUILD $@"
	@go build ${GOFLAG} ./tools/openedge-consistency

install: openedge
	install -d -m 0755 ${PREFIX}/bin
	install -m 0755 openedge ${PREFIX}/bin/
	tar cf - -C example/docker etc var | tar xvf - -C ${PREFIX}/

uninstall:
	rm -f ${PREFIX}/bin/openedge
	rm -rf ${PREFIX}/etc/openedge
	rm -rf ${PREFIX}/var/db/openedge
	rmdir ${PREFIX}/var/db
	rmdir ${PREFIX}/var
	rmdir ${PREFIX}/etc
	rmdir ${PREFIX}/bin
	rmdir ${PREFIX}

install-native: openedge package
	install -d -m 0755 ${PREFIX}/bin
	install -m 0755 openedge ${PREFIX}/bin/
	install -d -m 0755 ${PREFIX}/var/db/openedge/agent-bin
	tar xzvf openedge-agent/package.tar.gz -C ${PREFIX}/var/db/openedge/agent-bin
	install -d -m 0755 ${PREFIX}/var/db/openedge/localhub-bin
	tar xzvf openedge-hub/package.tar.gz -C ${PREFIX}/var/db/openedge/localhub-bin
	install -d -m 0755 ${PREFIX}/var/db/openedge/localfunc-bin
	tar xzvf openedge-function/package.tar.gz -C ${PREFIX}/var/db/openedge/localfunc-bin
	install -d -m 0755 ${PREFIX}/var/db/openedge/localfunc-python27
	tar xzvf openedge-function-python27/package.tar.gz -C ${PREFIX}/var/db/openedge/localfunc-python27
	install -d -m 0755 ${PREFIX}/var/db/openedge/bridge-mqtt-bin
	tar xzvf openedge-remote-mqtt/package.tar.gz -C ${PREFIX}/var/db/openedge/bridge-mqtt-bin
	tar cf - -C example/native etc var | tar xvf - -C ${PREFIX}/

uninstall-native:
	rm -f ${PREFIX}/bin/openedge
	rm -rf ${PREFIX}/etc/openedge
	rm -rf ${PREFIX}/var/db/openedge
	rmdir ${PREFIX}/var/db
	rmdir ${PREFIX}/var
	rmdir ${PREFIX}/etc
	rmdir ${PREFIX}/bin
	rmdir ${PREFIX}

.PHONY: clean
clean:
	rm -f openedge
	make -C openedge-agent clean
	make -C openedge-hub clean
	make -C openedge-function clean
	make -C openedge-function-python27 clean
	make -C openedge-remote-mqtt clean
	rm -f pubsub openedge-consistency

rebuild: clean all

pb: protobuf

protobuf:
	@echo "If protoc not installed, please get it from https://github.com/protocolbuffers/protobuf/releases"
	# protoc -Imodule/function/runtime --cpp_out=openedge-function-runtime-cxx --grpc_out=openedge-function-runtime-cxx --plugin=protoc-gen-grpc=`which grpc_cpp_plugin` openedge_function_runtime.proto
	protoc -Imodule/function/runtime --go_out=plugins=grpc:module/function/runtime openedge_function_runtime.proto
	python -m grpc_tools.protoc -Imodule/function/runtime --python_out=openedge-function-runtime-python27 --grpc_python_out=openedge-function-runtime-python27 openedge_function_runtime.proto

image:
	make -C openedge-hub image
	make -C openedge-function image
	make -C openedge-function-python27 image
	make -C openedge-remote-mqtt image
	make -C openedge-agent image

release:
	env GOOS=linux GOARCH=amd64 make image
	make clean
	# release linux 386
	env GOOS=linux GOARCH=386 make install PREFIX=__release_build/openedge-linux-386-$(VERSION)
	tar czf openedge-linux-386-$(VERSION).tar.gz -C __release_build/openedge-linux-386-$(VERSION) bin etc var
	tar cjf openedge-linux-386-$(VERSION).tar.bz2 -C __release_build/openedge-linux-386-$(VERSION) bin etc var
	make uninstall clean PREFIX=__release_build/openedge-linux-386-$(VERSION)
	# release linux amd64
	env GOOS=linux GOARCH=amd64 make install PREFIX=__release_build/openedge-linux-amd64-$(VERSION)
	tar czf openedge-linux-amd64-$(VERSION).tar.gz -C __release_build/openedge-linux-amd64-$(VERSION) bin etc var
	tar cjf openedge-linux-amd64-$(VERSION).tar.bz2 -C __release_build/openedge-linux-amd64-$(VERSION) bin etc var
	make uninstall clean PREFIX=__release_build/openedge-linux-amd64-$(VERSION)
	# release linux arm
	env GOOS=linux GOARCH=arm make install PREFIX=__release_build/openedge-linux-arm-$(VERSION)
	tar czf openedge-linux-arm-$(VERSION).tar.gz -C __release_build/openedge-linux-arm-$(VERSION) bin etc var
	tar cjf openedge-linux-arm-$(VERSION).tar.bz2 -C __release_build/openedge-linux-arm-$(VERSION) bin etc var
	make uninstall clean PREFIX=__release_build/openedge-linux-arm-$(VERSION)
	# release linux arm64
	env GOOS=linux GOARCH=arm64 make install PREFIX=__release_build/openedge-linux-arm64-$(VERSION)
	tar czf openedge-linux-arm64-$(VERSION).tar.gz -C __release_build/openedge-linux-arm64-$(VERSION) bin etc var
	tar cjf openedge-linux-arm64-$(VERSION).tar.bz2 -C __release_build/openedge-linux-arm64-$(VERSION) bin etc var
	make uninstall clean PREFIX=__release_build/openedge-linux-arm64-$(VERSION)
	# release darwin amd64
	env GOOS=darwin GOARCH=amd64 make all
	make install PREFIX=__release_build/openedge-darwin-amd64-$(VERSION)
	tar czf openedge-darwin-amd64-$(VERSION).tar.gz -C __release_build/openedge-darwin-amd64-$(VERSION) bin etc var
	tar cjf openedge-darwin-amd64-$(VERSION).tar.bz2 -C __release_build/openedge-darwin-amd64-$(VERSION) bin etc var
	make uninstall PREFIX=__release_build/openedge-darwin-amd64-$(VERSION)
	make install-native PREFIX=__release_build/openedge-darwin-amd64-$(VERSION)-native
	tar czf openedge-darwin-amd64-$(VERSION)-native.tar.gz -C __release_build/openedge-darwin-amd64-$(VERSION)-native bin etc var
	tar cjf openedge-darwin-amd64-$(VERSION)-native.tar.bz2 -C __release_build/openedge-darwin-amd64-$(VERSION)-native bin etc var
	make uninstall-native PREFIX=__release_build/openedge-darwin-amd64-$(VERSION)-native
	make clean
	# at last
	rmdir __release_build
