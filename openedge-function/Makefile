all: openedge-function

openedge-function:
	@echo "build ${GOFLAG} $@"
	@go build ${GOFLAG} .

openedge-function-image:
	@echo "build ${GOFLAG} $@"
	docker build -t openedge-function:build -f Dockerfile ../..

.PHONY: clean
clean:
	rm -f openedge-function

.PHONY: rebuild
rebuild: clean all
