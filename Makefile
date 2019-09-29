PREFIX?=/usr/local
MODE?=docker

GIT_REV:=git-$(shell git rev-parse --short HEAD)
GIT_TAG:=$(shell git tag --contains HEAD)
VERSION:=$(if $(GIT_TAG),$(GIT_TAG),$(GIT_REV))
CHANGES:=$(if $(shell git status -s),true,false)

GO_OS?=$(shell go env GOOS)
GO_ARCH?=$(shell go env GOARCH)
GO_ARM?=$(shell go env GOARM)
GO_FLAGS?=-ldflags "-X 'github.com/baetyl/baetyl/cmd.Revision=$(GIT_REV)' -X 'github.com/baetyl/baetyl/cmd.Version=$(VERSION)'"
GO_TEST_FLAGS?=
GO_TEST_PKGS?=$(shell go list ./... | grep -v baetyl-video-infer)

MODULES?=agent hub timer remote-mqtt # function-manager function-python function-node
ifndef PLATFORMS
	PLATFORMS:=$(if $(GO_ARM),$(GO_OS)/$(GO_ARCH)/$(GO_ARM),$(GO_OS)/$(GO_ARCH))
	ifeq ($(GO_OS),darwin)
		PLATFORMS+=linux/amd64
		XPLATFORMS:=linux/amd64
	endif
else
	ifeq ($(PLATFORMS),all)
		override PLATFORMS:=linux/amd64 linux/arm64 linux/ppc64le linux/s390x linux/386 linux/arm/v7 linux/arm/v6
	endif
	XPLATFORMS:=$(shell echo $(PLATFORMS) | sed 's: :,:g')
endif

OUTPUT:=output
OUTPUT_BINS:=$(PLATFORMS:%=$(OUTPUT)/%/baetyl/bin/baetyl)
OUTPUT_MODS:=$(MODULES:%=baetyl-%)
IMAGE_MODS:=$(MODULES:%=image/baetyl-%) # a little tricky to add prefix 'image/' in order to distinguish from OUTPUT_MODS
NATIVE_MODS:=$(MODULES:%=native/baetyl-%) # a little tricky to add prefix 'native/' in order to distinguish from OUTPUT_MODS

SRC=$(wildcard *.go) $(shell find cmd master logger sdk protocol utils -type f -name '*.go')

.PHONY: all $(OUTPUT_MODS)
all: $(OUTPUT_BINS) $(OUTPUT_MODS)

$(OUTPUT_BINS): $(SRC)
ifeq ($(CHANGES),true)
	$(error "Please commit or discard local changes")
endif
	@echo "BUILD $@"
	@mkdir -p $(dir $@)
	@$(shell echo $(patsubst $(OUTPUT)/%/baetyl/bin/baetyl,%,$@)  | sed 's:/v:/:g' | awk -F '/' '{print "CGO_ENABLED=0 GOOS="$$1" GOARCH="$$2" GOARM="$$3" go build"}') -o $@ ${GO_FLAGS} .

$(OUTPUT_MODS):
	@make -C $@

.PHONY: image $(IMAGE_MODS)
image: $(IMAGE_MODS) 

$(IMAGE_MODS):
	@make -C $(notdir $@) image

.PHONY: rebuild
rebuild: clean all

.PHONY: test
test:
	@go test ${GO_TEST_FLAGS} -coverprofile=coverage.out ${GO_TEST_PKGS}
	go tool cover -func=coverage.out | grep total

.PHONY: install $(NATIVE_MODS)
install: all
	@install -d -m 0755 ${PREFIX}/bin
	@install -m 0755 $(OUTPUT)/$(if $(GO_ARM),$(GO_OS)/$(GO_ARCH)/$(GO_ARM),$(GO_OS)/$(GO_ARCH))/baetyl/bin/baetyl ${PREFIX}/bin/
ifeq ($(MODE),native)
	make $(NATIVE_MODS)
endif
	@tar cf - -C example/$(MODE) etc var | tar xvf - -C ${PREFIX}/

$(NATIVE_MODS):
	install -d -m 0755 ${PREFIX}/var/db/baetyl/$(notdir $@)/bin
	install -m 0755 $(OUTPUT)/$(if $(GO_ARM),$(GO_OS)/$(GO_ARCH)/$(GO_ARM),$(GO_OS)/$(GO_ARCH))/$(notdir $@)/bin/$(notdir $@) ${PREFIX}/var/db/baetyl/$(notdir $@)/bin/
	install -m 0755 $(OUTPUT)/$(if $(GO_ARM),$(GO_OS)/$(GO_ARCH)/$(GO_ARM),$(GO_OS)/$(GO_ARCH))/$(notdir $@)/package.yml ${PREFIX}/var/db/baetyl/$(notdir $@)/

.PHONY: uninstall
uninstall:
	@-rm -f ${PREFIX}/bin/baetyl
	@-rm -rf ${PREFIX}/etc/baetyl
	@-rm -rf ${PREFIX}/var/db/baetyl
	@-rm -rf ${PREFIX}/var/log/baetyl
	@-rm -rf ${PREFIX}/var/run/baetyl
	@-rmdir ${PREFIX}/bin
	@-rmdir ${PREFIX}/etc
	@-rmdir ${PREFIX}/var/db
	@-rmdir ${PREFIX}/var/log
	@-rmdir ${PREFIX}/var/run
	@-rmdir ${PREFIX}/var
	@-rmdir ${PREFIX}

.PHONY: generate
generate:
	go generate ./...

.PHONY: clean
clean:
	@-rm -rf $(OUTPUT)