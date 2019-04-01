TAG?=latest
SRC=$(shell find . -type f -name '*.go')

all: openedge-function-manager package.zip

openedge-function-manager: $(SRC)
	@echo "BUILD $@"
	@go build ${GOFLAG} .

package.zip: openedge-function-manager
	install -d -m 0755 __package/bin
	cp $< __package/bin
	cp package.yml __package/
	cd __package && zip -q -r ../package.zip bin/ package.yml
	rm -rf __package

image: openedge-function-manager
	@echo "IMAGE $<"
	docker build -t $(IMAGE_PREFIX)$<$(IMAGE_SUFFIX):$(TAG) .

.PHONY: clean
clean:
	rm -f openedge-function-manager package.zip

.PHONY: rebuild
rebuild: clean all
