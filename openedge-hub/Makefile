all: openedge-hub

openedge-hub:
	@echo "build ${GOFLAG} $@"
	@env GOOS=$(GOOS) GOARCH=$(GOARCH) go build ${GOFLAG} -o openedge-hub$(TAG) .

openedge-hub-image:
	@echo "build ${GOFLAG} $@"
	docker build -t openedge-hub:build -f Dockerfile ../..

.PHONY: clean
clean:
	rm -f openedge-hub

rebuild: clean all