all: openedge-function

openedge-function:
	@echo "GO $@"
	@go build ${RACE} .

.PHONY: clean
clean:
	rm -f openedge-function

rebuild: clean all

