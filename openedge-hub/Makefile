TAG?=latest

all: openedge-hub

openedge-hub:
	@echo "BUILD $@"
	@go build ${GOFLAG} .

image: openedge-hub
	@echo "IMAGE $<"
	docker build -t $(IMAGE_PREFIX)$<:$(TAG) .

.PHONY: clean
clean:
	rm -f openedge-hub

rebuild: clean all
