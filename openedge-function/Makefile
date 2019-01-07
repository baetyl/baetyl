all: openedge-function

openedge-function:
	@echo "build ${GOFLAG} $@"
	@go build ${GOFLAG} .

openedge-function-release:
	@echo "build ${GOFLAG} $@"
	@env GOOS=$(GOOS) GOARCH=$(GOARCH) go build ${GOFLAG} -o ../output/openedge-$(GOOS)-$(GOARCH)/openedge-function .

openedge-function-image:
	@echo "build ${GOFLAG} $@"
	docker build -t openedge-function:build -f Dockerfile ../..

.PHONY: clean
clean:
	rm -f openedge-function

.PHONY: rebuild
rebuild: clean all
