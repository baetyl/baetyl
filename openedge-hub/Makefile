all: openedge-hub

openedge-hub:
	@echo "build ${GOFLAG} $@"
	@go build ${GOFLAG} .

openedge-hub-image:
	@echo "build ${GOFLAG} $@"
	docker build -t openedge-hub:build -f Dockerfile ../..

.PHONY: clean
clean:
	rm -f openedge-hub

rebuild: clean all
