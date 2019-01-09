all: openedge-hub

openedge-hub:
	@echo "build ${GOFLAG} $@"
	@go build ${GOFLAG} .

openedge-hub-release:
	@echo "build ${GOFLAG} $@"
	@env GOOS=$(GOOS) GOARCH=$(GOARCH) go build ${GOFLAG} -o ../output/openedge-$(GOOS)-$(GOARCH)/openedge-hub .

openedge-hub-image:
	@echo "build ${GOFLAG} $@"
	docker build -t openedge-hub:build -f Dockerfile ../..

.PHONY: clean
clean:
	rm -f openedge-hub

rebuild: clean all
