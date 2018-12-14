all: openedge-remote-mqtt

openedge-remote-mqtt:
	@echo "GO $@"
	@go build ${RACE} .

.PHONY: clean
clean:
	rm -f openedge-remote-mqtt

rebuild: clean all

