TAG?=latest
SRC=$(shell find . -type f -name '*.go')

all: openedge-function package.tar.gz

openedge-function: $(SRC)
	@echo "BUILD $@"
	@go build ${GOFLAG} .

package.tar.gz: openedge-function
	install -d -m 0755 __package/bin
	cp $< __package/bin
	cp package.yml __package/
	tar czvf $@ -C __package bin package.yml
	rm -rf __package

image: openedge-function
	@echo "IMAGE $<"
	docker build -t $(IMAGE_PREFIX)$<:$(TAG) .

.PHONY: clean
clean:
	rm -f openedge-function package.tar.gz

.PHONY: rebuild
rebuild: clean all
