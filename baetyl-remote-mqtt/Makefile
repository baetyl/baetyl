TAG?=latest
SRC=$(wildcard *.go)

all: baetyl-remote-mqtt package.zip

baetyl-remote-mqtt: $(SRC)
	@echo "BUILD $@"
	@CGO_ENABLED=0 go build ${GOFLAG} .

package.zip: baetyl-remote-mqtt
	install -d -m 0755 __package/bin
	cp $< __package/bin
	cp package.yml __package/
	cd __package && zip -q -r ../package.zip bin/ package.yml
	rm -rf __package

image: baetyl-remote-mqtt
	@echo "IMAGE $<"
	docker build -t $(IMAGE_PREFIX)$<:$(TAG) .

.PHONY: clean
clean:
	rm -f baetyl-remote-mqtt package.zip

rebuild: clean all
