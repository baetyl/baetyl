TAG?=latest

all: openedge-agent-bie

openedge-agent-bie:
	@echo "BUILD $@"
	@go build ${GOFLAG} .

image: openedge-agent-bie
	@echo "IMAGE $<"
	docker build -t $(IMAGE_PREFIX)$<:$(TAG) .

.PHONY: clean
clean:
	rm -f openedge-agent-bie

.PHONY: rebuild
rebuild: clean all
