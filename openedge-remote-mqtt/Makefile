TAG?=latest
SRC=$(wildcard *.go)

all: openedge-remote-mqtt package.zip

openedge-remote-mqtt: $(SRC)
	@echo "BUILD $@"
	@CGO_ENABLED=0 go build ${GOFLAG} .

package.zip: openedge-remote-mqtt
	install -d -m 0755 __package/bin
	cp $< __package/bin
	cp package.yml __package/
	cd __package && zip -q -r ../package.zip bin/ package.yml
	rm -rf __package

image: openedge-remote-mqtt
	@echo "IMAGE $<"
	docker build -t $(IMAGE_PREFIX)$<:$(TAG)$(IMAGE_SUFFIX) .

.PHONY: clean
clean:
	rm -f openedge-remote-mqtt package.zip

rebuild: clean all
