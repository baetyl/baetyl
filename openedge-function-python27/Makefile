TAG?=latest
PYSRC=$(shell find . -type f -name '*.py')

all: package.zip

package.zip: $(PYSRC)
	install -d -m 0755 __package/bin
	cp $^ __package/bin
	chmod +x __package/bin/openedge-function-python27.py
	cp package.yml __package/
	cd __package && zip -q -r ../package.zip bin/ package.yml
	rm -rf __package

package.tar.gz: $(PYSRC)
	install -d -m 0755 __package/bin
	cp $^ __package/bin
	chmod +x __package/bin/openedge-function-python27.py
	cp package.yml __package/
	tar czvf $@ -C __package bin package.yml
	rm -rf __package

image:
	@echo "IMAGE openedge-function-python27"
	docker build -t $(IMAGE_PREFIX)openedge-function-python27$(IMAGE_SUFFIX):$(TAG) .

.PHONY: clean
clean:
	rm -f package.zip
	rm -f *.pyc

.PHONY: rebuild
rebuild: clean all
