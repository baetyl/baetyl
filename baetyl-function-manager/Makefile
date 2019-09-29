MODULE:=baetyl-function-manager

GIT_REV?=git-$(shell git rev-parse --short HEAD)
GIT_TAG?=$(shell git tag --contains HEAD)
VERSION?=$(if $(GIT_TAG),$(GIT_TAG),$(GIT_REV))
CHANGES?=$(if $(shell git status -s),true,false)

IMAGE_REPO?=
IMAGE_PUSH?=
ifndef PLATFORMS
	GO_OS?=$(shell go env GOOS)
	GO_ARCH?=$(shell go env GOARCH)
	GO_ARM?=$(shell go env GOARM)
	PLATFORMS:=$(if $(GO_ARM),$(GO_OS)/$(GO_ARCH)/$(GO_ARM),$(GO_OS)/$(GO_ARCH))
	ifeq ($(GO_OS),darwin)
		PLATFORMS+=linux/amd64
		IMAGE_PLATFORMS=linux/amd64
	endif
else
	IMAGE_PLATFORMS=$(shell echo $(PLATFORMS) | sed 's: :,:g')
endif

OUTPUT?=../output
OUTPUT_MODS:=$(PLATFORMS:%=$(OUTPUT)/%/$(MODULE))
OUTPUT_BINS:=$(OUTPUT_MODS:%=%/bin/$(MODULE))
OUTPUT_PKGS:=$(OUTPUT_MODS:%=%/package.zip) # TODO: switch to tar

# TODO: to use VPATH
SRC=$(wildcard *.go)
YML=$(wildcard *.yml)

.PHONY: all
all: $(OUTPUT_BINS) $(OUTPUT_PKGS)

$(OUTPUT_BINS): $(SRC)
ifeq ($(CHANGES),true)
	$(error "Please commit or discard local changes")
endif
	@echo "BUILD $@"
	@mkdir -p $(dir $@)
	@$(shell echo $(patsubst $(OUTPUT)/%/$(MODULE)/bin/$(MODULE),%,$@)  | sed 's:/v:/:g' | awk -F '/' '{print "CGO_ENABLED=0 GOOS="$$1" GOARCH="$$2" GOARM="$$3" go build"}') -o $@ ${GO_FLAGS} .

$(OUTPUT_PKGS): $(OUTPUT_BINS) $(YML)
	@echo "PACKAGE $@"
	@install -m 0755 package.yml $(dir $@)
	@cd $(dir $@) && zip -q -r package.zip bin package.yml

.PHONY: image
image: $(OUTPUT_BINS)
	@echo "BUILDX: $(IMAGE_REPO)$(MODULE):$(VERSION)"
	@-docker buildx create --name baetyl
	@docker buildx use baetyl
	docker buildx build $(IMAGE_PUSH) --platform $(IMAGE_PLATFORMS) -t $(IMAGE_REPO)$(MODULE):$(VERSION) -f Dockerfile --no-cache ../output

.PHONY: clean $(OUTPUT_MODS)
clean: $(OUTPUT_MODS)

$(OUTPUT_MODS):
	@-rm -rf $@

.PHONY: rebuild
rebuild: clean all
