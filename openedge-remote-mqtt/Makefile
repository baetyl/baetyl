TAG?=latest

all: openedge-remote-mqtt

openedge-remote-mqtt:
	@echo "BUILD $@"
	@go build ${GOFLAG} .

image: openedge-remote-mqtt
	@echo "IMAGE $<"
	docker build -t $(IMAGE_PREFIX)$<:$(TAG) .

.PHONY: clean
clean:
	rm -f openedge-remote-mqtt

rebuild: clean all
