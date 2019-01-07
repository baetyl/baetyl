all: modify release-openedge release-images

version_info = "package module\n\n// Version the version of this binary\nconst Version = \"$(VERSION)\""

module_yml = "version: V0\nmodules:\n\t- name: localhub\n\t\tentry: '$(REPOSITORY)openedge-hub:$(VERSION)'\n\t\texpose:\n\t\t\t- 1883:1883\n\t- name: localhub\n\t\tentry: '$(REPOSITORY)openedge-function:$(VERSION)'\n\t\tresources:\n\t\t\tcpu:\n\t\t\t\tcpus: 1\n\t\t\tmemory:\n\t\t\t\tlimit: 500m"

modify:
	@echo $(module_yml) > ../example/docker/var/db/openedge/module/module.yml
	@echo $(version_info) > ../module/version.go

release-openedge:
	rm -rf ../output
	for ARCH in '386' 'amd64' 'arm' 'arm64'; do \
		make multi-openedge-builder GOOS=linux GOARCH=$$ARCH VERSION=$(VERSION); \
	done
	make multi-openedge-builder GOOS=darwin GOARCH=amd64 VERSION=$(VERSION)
	rm -rf ../output

release-images:
	docker build -t openedge-release:build -f Dockerfile-release ../..
	for ARCH in '386' 'amd64' 'arm' 'arm64'; do \
		make multi-images-builder GOOS=linux GOARCH=$$ARCH; \
	done
	docker rmi openedge-release:build

multi-openedge-builder:
	make openedge-release GOOS=$(GOOS) GOARCH=$(GOARCH)
	tar cf - -C ../example/docker . | tar xf - -C ../output/openedge-$(GOOS)-$(GOARCH)
	tar czvf ../openedge-$(GOOS)-$(GOARCH)-$(VERSION).tar.gz -C ../output/openedge-$(GOOS)-$(GOARCH) .

multi-modules-builder:
	for ARCH in '386' 'amd64' 'arm' 'arm64'; do \
		make -C ../openedge-hub openedge-hub-release GOOS=linux GOARCH=$$ARCH; \
		make -C ../openedge-function openedge-function-release GOOS=linux GOARCH=$$ARCH; \
		make -C ../openedge-remote-mqtt openedge-remote-mqtt-release GOOS=linux GOARCH=$$ARCH; \
	done

multi-images-builder:
	@echo "build ${GOFLAG} $@"
	docker build --build-arg GOOS=$(GOOS) --build-arg GOARCH=$(GOARCH) -t $(REPOSITORY)openedge-hub-$(GOOS)-$(GOARCH):$(VERSION) -f Dockerfile-hub .
	docker build --build-arg GOOS=$(GOOS) --build-arg GOARCH=$(GOARCH) -t $(REPOSITORY)openedge-function-$(GOOS)-$(GOARCH):$(VERSION) -f Dockerfile-function .
	docker build --build-arg GOOS=$(GOOS) --build-arg GOARCH=$(GOARCH) -t $(REPOSITORY)openedge-remote-mqtt-$(GOOS)-$(GOARCH):$(VERSION) -f Dockerfile-remote-mqtt .

openedge-release:
	@echo "build ${GOFLAG} $@"
	@env GOOS=$(GOOS) GOARCH=$(GOARCH) go build ${GOFLAG} -o ../output/openedge-$(GOOS)-$(GOARCH)/openedge ../