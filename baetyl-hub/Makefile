APP:=baetyl-hub

GIT_REV:=git-$(shell git rev-parse --short HEAD)
GIT_TAG:=$(shell git tag --contains HEAD)
VERSION:=$(if $(GIT_TAG),$(GIT_TAG),$(GIT_REV))
REGISTRY?=

.PHONY: all
all: $(APP)

.PHONY: clean
clean:
	@rm -f ${APP}

.PHONY: rebuild
rebuild: clean all

.PHONY: image
image: $(APP)
	@echo "BUILDX: $(REGISTRY)$(APP):$(VERSION)"
	docker build -t $(REGISTRY)$(APP):$(VERSION) -f Dockerfile .

$(APP):
	@echo "BUILD $@"
	CGO_ENABLED=0 go build -o $@ ./cmd
