PREFIX?=/usr/local
MODE?=docker
MODULES?=agent hub timer remote-mqtt function-manager function-node8 function-python3 function-python2
SRC_FILES:=$(shell find main.go cmd master logger sdk protocol utils -type f -name '*.go') # TODO use vpath
PLATFORM_ALL:=darwin/amd64 linux/amd64 linux/arm64 linux/386 linux/arm/v7 linux/arm/v6 linux/arm/v5 linux/ppc64le linux/s390x

GIT_REV:=git-$(shell git rev-parse --short HEAD)
GIT_TAG:=$(shell git tag --contains HEAD)
VERSION:=$(if $(GIT_TAG),$(GIT_TAG),$(GIT_REV))
# CHANGES:=$(if $(shell git status -s),true,false)

GO_OS:=$(shell go env GOOS)
GO_ARCH:=$(shell go env GOARCH)
GO_ARM:=$(shell go env GOARM)
GO_FLAGS?=-ldflags "-X 'github.com/baetyl/baetyl/cmd.Revision=$(GIT_REV)' -X 'github.com/baetyl/baetyl/cmd.Version=$(VERSION)'"
GO_FLAGS_STATIC=-ldflags '-X "github.com/baetyl/baetyl/cmd.Revision=$(GIT_REV)" -X "github.com/baetyl/baetyl/cmd.Version=$(VERSION)"  -linkmode external -w -extldflags "-static"'
GO_TEST_FLAGS?=-race -short -covermode=atomic -coverprofile=coverage.out
GO_TEST_PKGS?=$(shell go list ./... | grep -v baetyl-video-infer)

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

OUTPUT:=output
OUTPUT_DIRS:=$(PLATFORMS:%=$(OUTPUT)/%/baetyl)
OUTPUT_BINS:=$(OUTPUT_DIRS:%=%/bin/baetyl)
OUTPUT_PKGS:=$(OUTPUT_DIRS:%=%/baetyl-$(VERSION).zip) # TODO: switch to tar

OUTPUT_MODS:=$(MODULES:%=baetyl-%)
IMAGE_MODS:=$(MODULES:%=image/baetyl-%) # a little tricky to add prefix 'image/' in order to distinguish from OUTPUT_MODS
NATIVE_MODS:=$(MODULES:%=native/baetyl-%) # a little tricky to add prefix 'native/' in order to distinguish from OUTPUT_MODS

.PHONY: all $(OUTPUT_MODS)
all: baetyl $(OUTPUT_MODS)

baetyl: $(OUTPUT_BINS) $(OUTPUT_PKGS)

$(OUTPUT_BINS): $(SRC_FILES)
	@echo "BUILD $@"
	@mkdir -p $(dir $@)
	@# baetyl failed to collect cpu related data on darwin if set 'CGO_ENABLED=0' in compilation
	@$(shell echo $(@:$(OUTPUT)/%/baetyl/bin/baetyl=%)  | sed 's:/v:/:g' | awk -F '/' '{print "GOOS="$$1" GOARCH="$$2" GOARM="$$3" go build"}') -o $@ ${GO_FLAGS} .

$(OUTPUT_PKGS):
	@echo "PACKAGE $@"
	@cd $(dir $@) && zip -q -r $(notdir $@) bin 

$(OUTPUT_MODS):
	@${MAKE} -C $@

.PHONY: build
build: $(SRC_FILES)
	@echo "BUILD baetyl"
ifneq ($(GO_OS),darwin)
	@CGO_ENABLED=1 go build -o baetyl $(GO_FLAGS_STATIC) .
else
	@CGO_ENABLED=1 go build -o baetyl $(GO_FLAGS) .
endif

.PHONY: image $(IMAGE_MODS)
image: $(IMAGE_MODS) 

$(IMAGE_MODS):
	@${MAKE} -C $(notdir $@) image

.PHONY: rebuild
rebuild: clean all

.PHONY: test
test:
	@cd baetyl-function-node8 && npm install && cd -
	@cd baetyl-function-python2 && pip install -r requirements.txt && cd -
	@cd baetyl-function-python3 && pip3 install -r requirements.txt && cd -
	@go test ${GO_TEST_FLAGS} ${GO_TEST_PKGS}
	@go tool cover -func=coverage.out | grep total

.PHONY: install $(NATIVE_MODS)
install: all
	@install -d -m 0755 ${PREFIX}/bin
	@install -m 0755 $(OUTPUT)/$(if $(GO_ARM),$(GO_OS)/$(GO_ARCH)/$(GO_ARM),$(GO_OS)/$(GO_ARCH))/baetyl/bin/baetyl ${PREFIX}/bin/
ifeq ($(MODE),native)
	@${MAKE} $(NATIVE_MODS)
endif
	@tar cf - -C example/$(MODE) etc var | tar xvf - -C ${PREFIX}/

$(NATIVE_MODS):
	@install -d -m 0755 ${PREFIX}/var/db/baetyl/$(notdir $@)/bin
	@install -m 0755 $(OUTPUT)/$(if $(GO_ARM),$(GO_OS)/$(GO_ARCH)/$(GO_ARM),$(GO_OS)/$(GO_ARCH))/$(notdir $@)/bin/* ${PREFIX}/var/db/baetyl/$(notdir $@)/bin/
	@install -m 0755 $(OUTPUT)/$(if $(GO_ARM),$(GO_OS)/$(GO_ARCH)/$(GO_ARM),$(GO_OS)/$(GO_ARCH))/$(notdir $@)/package.yml ${PREFIX}/var/db/baetyl/$(notdir $@)/
	cd ${PREFIX}/var/db/baetyl/$(notdir $@)/bin && npm install

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
	@-rm -rf ./baetyl-function-node8/node_modules
	@-rm -rf $(OUTPUT)


.PHONY: fmt
fmt:
	go fmt  ./...
