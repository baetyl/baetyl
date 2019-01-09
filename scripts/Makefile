all: release-openedge release-images

release-openedge:
	@echo "package module\n\n// Version the version of this binary\nconst Version = \"$(VERSION)\"" > ../module/version.go
	for ARCH in '386' 'amd64' 'arm' 'arm64'; do \
		make multi-openedge-builder GOOS=linux GOARCH=$$ARCH VERSION=$(VERSION); \
	done
	# Only on Mac
	make multi-openedge-builder GOOS=darwin GOARCH=amd64 VERSION=$(VERSION)

release-images:
	docker build -t openedge-modules:release -f Dockerfile-release ../..
	for ARCH in '386' 'amd64' 'arm' 'arm64'; do \
		make multi-images-builder GOOS=linux GOARCH=$$ARCH; \
	done

multi-openedge-builder:
	make openedge-release GOOS=$(GOOS) GOARCH=$(GOARCH)
	tar cf - -C ../example/docker . | tar xf - -C ../output/openedge-$(GOOS)-$(GOARCH)
	tar czvf ../output/openedge-$(GOOS)-$(GOARCH)-$(VERSION).tar.gz -C ../output/openedge-$(GOOS)-$(GOARCH) .

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