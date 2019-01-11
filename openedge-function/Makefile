TAG?=latest

all: openedge-function

openedge-function:
	@echo "BUILD $@"
	@go build ${GOFLAG} .

image: openedge-function
	@echo "IMAGE $<"
	docker build -t $(IMAGE_PREFIX)$<:$(TAG) .

.PHONY: clean
clean:
	rm -f openedge-function

.PHONY: rebuild
rebuild: clean all
