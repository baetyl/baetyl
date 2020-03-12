GO_TEST_FLAGS?=-race -short -covermode=atomic -coverprofile=coverage.txt
GO_TEST_PKGS?=$(shell go list ./...)


.PHONY: test
test: format
	go test ${GO_TEST_FLAGS} ${GO_TEST_PKGS}

.PHONY: fmt format
fmt: format
format:
	go fmt ${GO_TEST_PKGS}

.PHONY: prepare
prepare:
	go env -w GONOPROXY=\*\*.baidu.com\*\*
	go env -w GOPROXY=https://goproxy.baidu.com
	go env -w GONOSUMDB=\*
	go mod download