TAG?=latest
PYSRC=$(shell find . -type f -name '*.py')

all: package.tar.gz

package.tar.gz: $(PYSRC)
	install -d -m 0755 __package/bin
	cp $^ __package/bin
	chmod +x __package/bin/openedge-function-python27.py
	cp package.yml __package/
	tar czvf $@ -C __package bin package.yml
	rm -rf __package

image:
	@echo "IMAGE $<"
	docker build -t $(IMAGE_PREFIX)$<(IMAGE_SUFFIX):$(TAG) .

.PHONY: clean
clean:
	rm -f package.tar.gz
	rm -f *.pyc

.PHONY: rebuild
rebuild: clean all
