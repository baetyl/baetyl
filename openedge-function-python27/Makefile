TAG?=latest
GOSRC=$(shell find . -type f -name '*.go')
PYSRC=$(shell find . -type f -name '*.py')

all: openedge-function-python27 package.tar.gz

openedge-function-python27: $(GOSRC)
	@echo "BUILD $@"
	@go build ${GOFLAG} .

package.tar.gz: openedge-function-python27 runtime $(PYSRC)
	install -d -m 0755 __package/bin
	cp $^ __package/bin
	chmod +x __package/bin/runtime
	cp package.yml __package/
	tar czvf $@ -C __package bin package.yml
	rm -rf __package

image: openedge-function-python27
	@echo "IMAGE openedge-function-python27"
	docker build -t $(IMAGE_PREFIX)openedge-function-python27:$(TAG) .

.PHONY: clean
clean:
	rm -f openedge-function-python27 package.tar.gz

.PHONY: rebuild
rebuild: clean all
