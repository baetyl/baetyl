TAG?=latest
SRC=$(wildcard *.go)

all: openedge-agent package.zip

openedge-agent: $(SRC)
	@echo "BUILD $@"
	@CGO_ENABLED=0 go build ${GOFLAG} .

package.zip: openedge-agent
	install -d -m 0755 __package/bin
	cp $< __package/bin
	cp package.yml __package/
	cd __package && zip -q -r ../package.zip bin/ package.yml
	rm -rf __package

image: openedge-agent
	@echo "IMAGE $<"
	docker build -t $(IMAGE_PREFIX)$<:$(TAG)$(IMAGE_SUFFIX) .

.PHONY: clean
clean:
	rm -f openedge-agent package.zip

.PHONY: rebuild
rebuild: clean all
