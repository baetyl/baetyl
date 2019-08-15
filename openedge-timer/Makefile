SRC=$(wildcard *.go)

all: openedge-timer package.zip

openedge-timer: $(SRC)
	@echo "BUILD $@"
	@CGO_ENABLED=0 go build ${GOFLAG} .

package.zip: openedge-timer
	install -d -m 0755 __package/bin
	cp $< __package/bin
	cp package.yml __package/
	cd __package && zip -q -r ../package.zip bin/ package.yml
	rm -rf __package

image: openedge-timer
	@echo "IMAGE $<"
	docker build -t $(IMAGE_PREFIX)/$<:$(VERSION)$(IMAGE_SUFFIX) .

.PHONY: clean
clean:
	rm -f openedge-timer package.zip

.PHONY: rebuild
rebuild: clean all
