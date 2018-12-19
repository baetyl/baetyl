all: openedge-function

openedge-function:
	@echo "GO $@ ${GOFLAG}"
	@go build ${GOFLAG} .

.PHONY: clean
clean:
	rm -f openedge-function

rebuild: clean all
