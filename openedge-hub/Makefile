all: openedge-hub

openedge-hub:
	@echo "GO $@"
	@go build ${RACE} .

.PHONY: clean
clean:
	rm -f openedge-hub

rebuild: clean all

