TAG?=latest
SRC=$(shell find . -type f -name '*.go')

all: openedge-hub package.zip

openedge-hub: $(SRC)
	@echo "BUILD $@"
	@CGO_ENABLED=0 go build ${GOFLAG} .

package.zip: openedge-hub
	install -d -m 0755 __package/bin
	cp $< __package/bin
	cp package.yml __package/
	cd __package && zip -q -r ../package.zip bin/ package.yml
	rm -rf __package

image: openedge-hub
	@echo "IMAGE $<"
	docker build -t $(IMAGE_PREFIX)$<:$(TAG)$(IMAGE_SUFFIX) .

.PHONY: clean
clean:
	rm -f openedge-hub package.zip

rebuild: clean all
