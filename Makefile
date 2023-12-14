HOMEDIR := $(shell pwd)
OUTDIR  := $(HOMEDIR)/output

MODULE:=baetyl
SRC_FILES:=$(shell find . -type f -name '*.go')
PLATFORM_ALL:=darwin/amd64 linux/amd64 linux/arm64 linux/arm/v7 windows/amd64

export DOCKER_CLI_EXPERIMENTAL=enabled

GIT_TAG:=$(shell git tag --contains HEAD)
GIT_REV:=git-$(shell git rev-parse --short HEAD)
VERSION:=$(if $(GIT_TAG),$(GIT_TAG),$(GIT_REV))

GO       = go
GO_MOD   = $(GO) mod
GO_ENV   = env CGO_ENABLED=0
GO_FLAGS:=-ldflags '-s -w -X "github.com/baetyl/baetyl-go/v2/utils.REVISION=$(GIT_REV)" -X "github.com/baetyl/baetyl-go/v2/utils.VERSION=$(VERSION)"'
GO_TEST_FLAGS:=-race -short -covermode=atomic -coverprofile=coverage.txt
GO_TEST_PKGS:=$(shell go list ./...)
GO_BUILD = $(GO_ENV) $(GO) build $(GO_FLAGS)
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
XPLATFORMS:=$(shell echo $(filter-out darwin/amd64 windows/amd64,$(PLATFORMS)) | sed 's: :,:g')

OUTPUT     :=output
OUTPUT_DIRS:=$(PLATFORMS:%=$(OUTPUT)/%/$(MODULE))
OUTPUT_BINS:=$(OUTPUT_DIRS:%=%/$(MODULE))
PKG_PLATFORMS := $(shell echo $(PLATFORMS) | sed 's:/:-:g')
OUTPUT_PKGS:=$(PKG_PLATFORMS:%=$(OUTPUT)/$(MODULE)_%_$(VERSION).zip)

.PHONY: all
all: build

.PHONY: build
build: $(OUTPUT_BINS)

$(OUTPUT_BINS): $(SRC_FILES)
	@echo "BUILD $@"
	@mkdir -p $(dir $@)
	@mkdir -p $(dir $@)page
	@cp program.yml $(dir $@)
	@cp res/* $(dir $@)page
	@$(shell echo $(@:$(OUTPUT)/%/$(MODULE)/$(MODULE)=%)  | sed 's:/v:/:g' | awk -F '/' '{print "GOOS="$$1" GOARCH="$$2" GOARM="$$3""}') $(GO_BUILD) -o $@ .

.PHONY: build
build: $(OUTPUT_BINS)

.PHONY: build-local
build-local: $(SRC_FILES)
	@echo "BUILD $(MODULE)"
	$(GO_MOD) tidy
	$(GO_BUILD) -o $(MODULE) .
	@chmod +x $(MODULE)

.PHONY: image
image:
	@echo "BUILDX: $(REGISTRY)$(MODULE):$(VERSION)"
	@-docker buildx create --name baetyl
	@docker buildx use baetyl
	@docker run --privileged --rm tonistiigi/binfmt --install all
	docker buildx build $(XFLAGS) --platform $(XPLATFORMS) -t $(REGISTRY)$(MODULE):$(VERSION) -f Dockerfile .

.PHONY: test
test: fmt
	@go test ${GO_TEST_FLAGS} ${GO_TEST_PKGS}
	@go tool cover -func=coverage.txt | grep total

.PHONY: fmt
fmt:
	go fmt ./...

package: build $(OUTPUT_PKGS)

$(OUTPUT_PKGS):
	@echo "PACKAGE $@"
	@cd $(OUTPUT)/$(shell echo $(@:$(OUTPUT)/$(MODULE)_%_$(VERSION).zip=%) | sed 's:-:/:g')/$(MODULE) && zip -q -r $(notdir $@) $(MODULE) program.yml page
	@$(shell if [ $(shell echo $(@:$(OUTPUT)/$(MODULE)_%_$(VERSION).zip=%) | sed 's:-:/:g') == "windows/amd64" ];then \
		cd $(OUTPUT)/windows/amd64/$(MODULE) && mv $(MODULE) $(MODULE).exe && zip -q -r $(MODULE)_windows-amd64_$(VERSION).zip $(MODULE).exe program.yml page ; fi)

.PHONY: clean
clean:
	rm -rf $(OUTDIR)
	rm -rf $(HOMEDIR)/$(MODULE)
