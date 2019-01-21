TAG?=latest
SRC=$(wildcard *.go)

all: openedge-remote-mqtt package.tar.gz

openedge-remote-mqtt: $(SRC)
	@echo "BUILD $@"
	@go build ${GOFLAG} .

package.tar.gz: openedge-remote-mqtt
	install -d -m 0755 __package/bin
	cp $< __package/bin
	cp package.yml __package/
	tar czvf $@ -C __package bin package.yml
	rm -rf __package

image: openedge-remote-mqtt
	@echo "IMAGE $<"
	docker build -t $(IMAGE_PREFIX)$<:$(TAG) .

.PHONY: clean
clean:
	rm -f openedge-remote-mqtt package.tar.gz

rebuild: clean all
