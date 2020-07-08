MODULE_CORE:=core
MODULE_INIT:=init
BIN_CORE:=baetyl-$(MODULE_CORE)
BIN_INIT:=baetyl-$(MODULE_INIT)
SRC_FILES:=$(shell find . -type f -name '*.go')
PLATFORM_ALL:=darwin/amd64 linux/amd64 linux/arm64 linux/arm/v7

export DOCKER_CLI_EXPERIMENTAL=enabled

GIT_TAG:=$(shell git tag --contains HEAD)
GIT_REV:=git-$(shell git rev-parse --short HEAD)
VERSION:=$(if $(GIT_TAG),$(GIT_TAG),$(GIT_REV))

GO_FLAGS:=-ldflags '-s -w -X "github.com/baetyl/baetyl-go/utils.REVISION=$(GIT_REV)" -X "github.com/baetyl/baetyl-go/utils.VERSION=$(VERSION)"'
GO_TEST_FLAGS:=-race -short -covermode=atomic -coverprofile=coverage.txt
GO_TEST_PKGS:=$(shell go list ./...)
ifndef PLATFORMS
	GO_OS:=$(shell go env GOOS)
	GO_ARCH:=$(shell go env GOARCH)
	GO_ARM:=$(shell go env GOARM)
	PLATFORMS:=$(if $(GO_ARM),$(GO_OS)/$(GO_ARCH)/$(GO_ARM),$(GO_OS)/$(GO_ARCH))
	ifeq ($(GO_OS),darwin)
		PLATFORMS+=linux/amd64
	endif
else ifeq ($(PLATFORMS),all)
	override PLATFORMS:=$(PLATFORM_ALL)
endif

REGISTRY:=
XFLAGS:=--load
XPLATFORMS:=$(shell echo $(filter-out darwin/amd64,$(PLATFORMS)) | sed 's: :,:g')

.PHONY: all
all: core init

.PHONY: core
core: $(SRC_FILES)
	@echo "BUILD $(BIN_CORE)"
	@env GO111MODULE=on GOPROXY=https://goproxy.cn CGO_ENABLED=0 go build -o $(BIN_CORE) $(GO_FLAGS) ./cmd/core

.PHONY: init
init: $(SRC_FILES)
	@echo "BUILD $(BIN_INIT)"
	@env GO111MODULE=on GOPROXY=https://goproxy.cn CGO_ENABLED=0 go build -o $(BIN_INIT) $(GO_FLAGS) ./cmd/initz

.PHONY: image
image: image-core image-init

.PHONY: image-core
image-core:
	@echo "BUILDX: $(REGISTRY)$(MODULE_CORE):$(VERSION)"
	@-docker buildx create --name baetyl
	@docker buildx use baetyl
	@docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
	docker buildx build $(XFLAGS) --platform $(XPLATFORMS) -t $(REGISTRY)$(MODULE_CORE):$(VERSION) -f Dockerfile .

.PHONY: image-init
image-init:
	@echo "BUILDX: $(REGISTRY)$(MODULE_INIT):$(VERSION)"
	@-docker buildx create --name baetyl
	@docker buildx use baetyl
	@docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
	docker buildx build $(XFLAGS) --platform $(XPLATFORMS) -t $(REGISTRY)$(MODULE_INIT):$(VERSION) -f init.Dockerfile .

.PHONY: test
test: fmt
	@go test ${GO_TEST_FLAGS} ${GO_TEST_PKGS}
	@go tool cover -func=coverage.txt | grep total

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: clean
clean:
	@rm -rf $(BIN_CORE)
	@rm -rf $(BIN_INIT)