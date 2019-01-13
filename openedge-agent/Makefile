TAG?=latest

all: openedge-agent

openedge-agent:
	@echo "BUILD $@"
	@go build ${GOFLAG} .

image: openedge-agent
	@echo "IMAGE $<"
	docker build -t $(IMAGE_PREFIX)$<:$(TAG) .

.PHONY: clean
clean:
	rm -f openedge-agent

.PHONY: rebuild
rebuild: clean all
