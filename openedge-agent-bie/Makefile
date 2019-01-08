all: openedge-agent-bie

openedge-agent-bie:
	@echo "build ${GOFLAG} $@"
	@go build ${GOFLAG} .

openedge-agent-bie-image:
	@echo "build ${GOFLAG} $@"
	docker build -t openedge-agent-bie:build -f Dockerfile ../..

.PHONY: clean
clean:
	rm -f openedge-agent-bie

rebuild: clean all
