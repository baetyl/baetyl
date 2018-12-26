all: openedge-remote-mqtt

openedge-remote-mqtt:
	@echo "build ${GOFLAG} $@"
	@env GOOS=$(GOOS) GOARCH=$(GOARCH) go build ${GOFLAG} -o openedge-remote-mqtt$(TAG) .

openedge-remote-mqtt-image:
	@echo "build ${GOFLAG} $@"
	docker build -t openedge-remote-mqtt:build -f Dockerfile ../..

.PHONY: clean
clean:
	rm -f openedge-remote-mqtt

rebuild: clean all