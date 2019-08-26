TAG?=latest
SRC=$(wildcard *.go)

all: baetyl-agent package.zip

baetyl-agent: $(SRC)
	@echo "BUILD $@"
	@CGO_ENABLED=0 go build ${GOFLAG} .

package.zip: baetyl-agent
	install -d -m 0755 __package/bin
	cp $< __package/bin
	cp package.yml __package/
	cd __package && zip -q -r ../package.zip bin/ package.yml
	rm -rf __package

image: baetyl-agent
	@echo "IMAGE $<"
	docker build -t $(IMAGE_PREFIX)$<:$(TAG) .

.PHONY: clean
clean:
	rm -f baetyl-agent package.zip

.PHONY: rebuild
rebuild: clean all
