all: openedge-hub

openedge-hub:
	@echo "GO $@ ${GOFLAG}"
	@go build ${GOFLAG} .

.PHONY: clean
clean:
	rm -f openedge-hub

rebuild: clean all
