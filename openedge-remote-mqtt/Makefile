all: openedge-remote-mqtt

openedge-remote-mqtt:
	@echo "GO $@ ${GOFLAG}"
	@go build ${GOFLAG} .

.PHONY: clean
clean:
	rm -f openedge-remote-mqtt

rebuild: clean all
