VERSION?=latest
PYSRC27=$(shell find . -type f -name '*.py' -not -name '*36*')
PYSRC36=$(shell find . -type f -name '*.py' -not -name '*27*')

all: package27.zip package36.zip

package27.zip: $(PYSRC27)
	@echo "PACKAGE $<"
	@install -d -m 0755 __package/bin
	@cp $^ __package/bin
	@chmod +x __package/bin/function-python27.py
	@cp package27.yml __package/package.yml
	@cd __package && zip -q -r ../package27.zip bin/ package.yml
	@rm -rf __package

package36.zip: $(PYSRC36)
	@echo "PACKAGE $<"
	@install -d -m 0755 __package/bin
	@cp $^ __package/bin
	@chmod +x __package/bin/function-python36.py
	@cp package36.yml __package/package.yml
	@cd __package && zip -q -r ../package36.zip bin/ package.yml
	@rm -rf __package

buildx:
	@echo "BUILDX baetyl-function-python27"
	@docker build -t $(DOCKER_REPO)baetyl-function-python27:$(VERSION) -f Dockerfile-python27 .
	@echo "BUILDX baetyl-function-python36"
	@docker build -t $(DOCKER_REPO)baetyl-function-python36:$(VERSION) -f Dockerfile-python36 .

buildx-devel:
	@echo "BUILDX baetyl-function-python27 devel"
	@docker build -t $(DOCKER_REPO)baetyl-function-python27:$(VERSION) -f Dockerfile-python27-devel .
	@echo "BUILDX baetyl-function-python36 devel"
	@docker build -t $(DOCKER_REPO)baetyl-function-python36:$(VERSION) -f Dockerfile-python36-devel .

.PHONY: clean
clean:
	rm -f package27.zip package36.zip
	rm -f *.pyc

.PHONY: rebuild
rebuild: clean all
