all: release-prepare release-openedge release-images

release-prepare:
	@echo "package module\n\n// Version the version of this binary\nconst Version = \"$(VERSION)\"" > ../module/version.go
	rm -rf ../output
	mkdir -p ../output/config
	tar cf - -C ../example/docker . | tar xf - -C ../output/config
	cat ../output/config/var/db/openedge/module/module.yml | sed 's|openedge-hub:build|$(REPOSITORY)openedge-hub:$(VERSION)|' | sed 's|openedge-function:build|$(REPOSITORY)openedge-function:$(VERSION)|' > module.yml
	mv module.yml ../output/config/var/db/openedge/module/

release-openedge:
	for ARCH in '386' 'amd64' 'arm' 'arm64'; do \
		make multi-openedge-builder GOOS=linux GOARCH=$$ARCH VERSION=$(VERSION); \
	done
	make multi-openedge-builder GOOS=darwin GOARCH=amd64 VERSION=$(VERSION)
	rm -rf ../output

release-images:
	docker build -t openedge-modules:release -f Dockerfile-release ../..
	for ARCH in '386' 'amd64' 'arm' 'arm64'; do \
		make multi-images-builder GOOS=linux GOARCH=$$ARCH; \
	done
	docker rmi openedge-modules:release

multi-openedge-builder:
	make openedge-release GOOS=$(GOOS) GOARCH=$(GOARCH)
	tar cf - -C ../output/config . | tar xf - -C ../output/openedge-$(GOOS)-$(GOARCH)
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