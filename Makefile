GOOS ?= $(shell go env GOOS)
-include Makefile.$(GOOS)

GIT_REV:=git-$(shell git rev-parse --short HEAD)
GIT_TAG:=$(shell git tag --contains HEAD)
VERSION:=$(if $(GIT_TAG),$(GIT_TAG),$(GIT_REV))

LDFLAGS?=\
	-X 'github.com/baetyl/baetyl.Revision=$(GIT_REV)' \
	-X 'github.com/baetyl/baetyl.Version=$(VERSION)' \
	-X 'github.com/baetyl/baetyl.DefaultPrefix=$(PREFIX)' \
	-X 'github.com/baetyl/baetyl.DefaultConfigPath=$(CONF_PATH)' \
	-X 'github.com/baetyl/baetyl.DefaultLoggerPath=$(LOG_PATH)' \
	-X 'github.com/baetyl/baetyl.DefaultDataPath=$(DATA_PATH)' \
	-X 'github.com/baetyl/baetyl.DefaultPidPath=$(PID_PATH)' \
	-X 'github.com/baetyl/baetyl.DefaultAPIAddress=$(API_ADDR)'
GO_TEST_FLAGS?=-race -short -covermode=atomic -coverprofile=coverage.out

.PHONY: all
all: $(APP)
	CGO_ENABLED=${CGO_ENABLED} CC=${CC} make -C baetyl-hub all

$(APP):
	@echo "BUILD $@"
	CGO_ENABLED=${CGO_ENABLED} CC=${CC} go build -ldflags "${LDFLAGS}" -o $@ ./cmd/baetyl

.PHONY: clean
clean:
	@-rm -f $(APP)
	make -C baetyl-hub clean

.PHONY: rebuild
rebuild: clean all

.PHONY: image
image: all
	make -C baetyl-hub image

.PHONY: test
test:
	@cd baetyl-function-node8 && npm install && cd -
	@cd baetyl-function-python2 && pip install -r requirements.txt && cd -
	@cd baetyl-function-python3 && pip3 install -r requirements.txt && cd -
	@go test ${GO_TEST_FLAGS} ${GO_TEST_PKGS}
	@go tool cover -func=coverage.out | grep total

.PHONY: fmt
fmt:
	go fmt  ./...
