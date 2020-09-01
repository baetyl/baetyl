MODULE:=baetyl
SRC_FILES:=$(shell find . -type f -name '*.go')
PLATFORM_ALL:=darwin/amd64 linux/amd64 linux/arm64 linux/arm/v7

export DOCKER_CLI_EXPERIMENTAL=enabled

GIT_TAG:=$(shell git tag --contains HEAD)
GIT_REV:=git-$(shell git rev-parse --short HEAD)
VERSION:=$(if $(GIT_TAG),$(GIT_TAG),$(GIT_REV))
ifeq ($(findstring race,$(BUILD_ARGS)),race)
VERSION:=$(VERSION)-race
endif

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

GO       := go
GO_MOD   := $(GO) mod
GO_ENV   := env GO111MODULE=on GOPROXY=https://goproxy.cn CGO_ENABLED=0
GO_FLAGS := $(BUILD_ARGS) -ldflags '-X "github.com/baetyl/baetyl-go/v2/utils.REVISION=$(GIT_REV)" -X "github.com/baetyl/baetyl-go/v2/utils.VERSION=$(VERSION)"'
ifeq ($(findstring race,$(BUILD_ARGS)),race)
GO_ENV   := env GO111MODULE=on GOPROXY=https://goproxy.cn CGO_ENABLED=1
GO_FLAGS := $(BUILD_ARGS) -ldflags '-s -w -X "github.com/baetyl/baetyl-go/v2/utils.REVISION=$(GIT_REV)" -X "github.com/baetyl/baetyl-go/v2/utils.VERSION=$(VERSION)"  -linkmode external -w -extldflags "-static"'
override PLATFORMS:= $(filter-out linux/arm/v7,$(PLATFORMS))
endif
GO_BUILD := $(GO_ENV) $(GO) build $(GO_FLAGS)
GOTEST   := $(GO) test
GOPKGS   := $$($(GO) list ./... | grep -vE "vendor")

REGISTRY:=
XFLAGS:=--load
XPLATFORMS:=$(shell echo $(filter-out darwin/amd64,$(PLATFORMS)) | sed 's: :,:g')

.PHONY: all
all: build

.PHONY: build
build: $(SRC_FILES)
	@echo "BUILD $(MODULE)"
	$(GO_BUILD) -o $(MODULE) .
	@chmod +x $(MODULE)

.PHONY: image
image:
	@echo "BUILDX: $(REGISTRY)$(MODULE):$(VERSION)"
	@-docker buildx create --name baetyl
	@docker buildx use baetyl
	@docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
	docker buildx build $(XFLAGS) --platform $(XPLATFORMS) -t $(REGISTRY)$(MODULE):$(VERSION) --build-arg BUILD_ARGS=$(BUILD_ARGS) -f Dockerfile .

.PHONY: test
test: fmt
	$(GOTEST) -race -short -covermode=atomic -coverprofile=coverage.txt $(GOPKGS)
	@go tool cover -func=coverage.txt | grep total

.PHONY: fmt
fmt:
	$(GO_MOD) tidy
	@go fmt ./...

.PHONY: clean
clean:
	@rm -rf $(MODULE)

.PHONY: package
package: build
	zip baetyl_$(GO_OS)_$(GO_ARCH).zip program.yml baetyl