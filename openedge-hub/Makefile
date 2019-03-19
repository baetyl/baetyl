TAG?=latest
SRC=$(shell find . -type f -name '*.go')

all: openedge-hub package.tar.gz

openedge-hub: $(SRC)
	@echo "BUILD $@"
	@go build ${GOFLAG} .

package.tar.gz: openedge-hub
	install -d -m 0755 __package/bin
	cp $< __package/bin
	cp package.yml __package/
	tar czvf $@ -C __package bin package.yml
	rm -rf __package

image: openedge-hub
	@echo "IMAGE $<"
	docker build -t $(IMAGE_PREFIX)$<$(IMAGE_SUFFIX):$(VERSION) .
	docker tag ${IMAGE_PREFIX}$<${IMAGE_SUFFIX}:$(VERSION) $(IMAGE_PREFIX)$<$(IMAGE_SUFFIX):$(TAG)

.PHONY: clean
clean:
	rm -f openedge-hub package.tar.gz

rebuild: clean all
