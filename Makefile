PREFIX?=/usr/local

all: openedge modules

modules: openedge-hub/openedge-hub openedge-function/openedge-function openedge-remote-mqtt/openedge-remote-mqtt

openedge:
	@echo "build ${GOFLAG} $@"
	@go build ${GOFLAG} .

openedge-release:
	@echo "build ${GOFLAG} $@"
	@env GOOS=$(GOOS) GOARCH=$(GOARCH) go build ${GOFLAG} -o output/openedge-$(GOOS)-$(GOARCH)/openedge .

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
	@echo "build ${GOFLAG} $@"
	@go build ${GOFLAG} ./tools/pubsub

openedge-consistency:
	@echo "build ${GOFLAG} $@"
	@go build ${GOFLAG} ./tools/openedge-consistency

install: all
	install -d -m 0755 ${PREFIX}/bin
	install -m 0755 openedge ${PREFIX}/bin/
	install -m 0755 openedge-hub/openedge-hub ${PREFIX}/bin/
	install -m 0755 openedge-function/openedge-function ${PREFIX}/bin/
	install -m 0755 openedge-remote-mqtt/openedge-remote-mqtt ${PREFIX}/bin/
	install -m 0755 openedge-function-runtime-python27/openedge_function_runtime_pb2.py ${PREFIX}/bin
	install -m 0755 openedge-function-runtime-python27/openedge_function_runtime_pb2_grpc.py ${PREFIX}/bin
	install -m 0755 openedge-function-runtime-python27/openedge_function_runtime_python27.py ${PREFIX}/bin
	tar cf - -C example/native etc var | tar xvf - -C ${PREFIX}/

uninstall:
	rm -f ${PREFIX}/bin/openedge
	rm -f ${PREFIX}/bin/openedge-hub
	rm -f ${PREFIX}/bin/openedge-function
	rm -f ${PREFIX}/bin/openedge-remote-mqtt
	rm -f ${PREFIX}/bin/openedge_function_runtime_pb2.py
	rm -f ${PREFIX}/bin/openedge_function_runtime_pb2.pyc
	rm -f ${PREFIX}/bin/openedge_function_runtime_pb2_grpc.py
	rm -f ${PREFIX}/bin/openedge_function_runtime_pb2_grpc.pyc
	rm -f ${PREFIX}/bin/openedge_function_runtime_python27.py
	rm -f ${PREFIX}/bin/openedge_function_runtime_python27.pyc
	rm -rf ${PREFIX}/var/log/openedge
	rm -rf ${PREFIX}/var/db/openedge
	rm -rf ${PREFIX}/etc/openedge
	rmdir ${PREFIX}/var/log
	rmdir ${PREFIX}/var/db
	rmdir ${PREFIX}/var
	rmdir ${PREFIX}/etc
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

pb: protobuf

protobuf:
	@echo "If protoc not installed, please get it from https://github.com/protocolbuffers/protobuf/releases"
	# protoc -Imodule/function/runtime --cpp_out=openedge-function-runtime-cxx --grpc_out=openedge-function-runtime-cxx --plugin=protoc-gen-grpc=`which grpc_cpp_plugin` openedge_function_runtime.proto
	protoc -Imodule/function/runtime --go_out=plugins=grpc:module/function/runtime openedge_function_runtime.proto
	python -m grpc_tools.protoc -Imodule/function/runtime --python_out=openedge-function-runtime-python27 --grpc_python_out=openedge-function-runtime-python27 openedge_function_runtime.proto

images: openedge-hub-image openedge-function-image openedge-remote-mqtt-image openedge-function-runtime-python27-image

openedge-hub-image:
	make -C openedge-hub openedge-hub-image

openedge-function-image:
	make -C openedge-function openedge-function-image

openedge-remote-mqtt-image:
	make -C openedge-remote-mqtt openedge-remote-mqtt-image

openedge-function-runtime-python27-image:
	make -C openedge-function-runtime-python27 openedge-function-runtime-python27-image

version_info = "package module\n\n// Version the version of this binary\nconst Version = \"$(VERSION)\""
release:
	@echo $(version_info) > module/version.go
	rm -rf output
	for ARCH in '386' 'amd64' 'arm' 'arm64'; do \
		make multi-builder GOOS=linux GOARCH=$$ARCH VERSION=$(VERSION); \
	done
	# only on mac os
	# make multi-build GOOS=darwin GOARCH=amd64
	rm -rf output

release-test:
	make multi-images-builder

multi-openedge-builder:
	make openedge-release GOOS=$(GOOS) GOARCH=$(GOARCH)
	tar cf - -C example/docker . | tar xf - -C output/openedge-$(GOOS)-$(GOARCH)
	tar czvf openedge-$(GOOS)-$(GOARCH)-$(VERSION).tar.gz -C output/openedge-$(GOOS)-$(GOARCH) .

multi-modules-builder:
	make -C openedge-hub openedge-hub-release GOOS=$(GOOS) GOARCH=$(GOARCH)
	make -C openedge-function openedge-function-release GOOS=$(GOOS) GOARCH=$(GOARCH)
	make -C openedge-remote-mqtt openedge-remote-mqtt-release GOOS=$(GOOS) GOARCH=$(GOARCH)

multi-images-builder:
	@echo "build ${GOFLAG} $@"
	docker build -t openedge-release:build -f releases/Dockerfile-release .
