PREFIX?=/usr/local
VERSION?=git-$(shell git rev-list HEAD|head -1|cut -c 1-6)
GOFLAG?=-ldflags "-X github.com/baidu/openedge/cmd.BuildTime=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X 'github.com/baidu/openedge/cmd.GoVersion=`go version`' -X 'github.com/baidu/openedge/cmd.Version=$(VERSION)' -X 'master.Version=$(VERSION)'"

all: openedge package

package: \
	openedge-hub/package.tar.gz \
	openedge-agent/package.tar.gz \
	openedge-remote-mqtt/package.tar.gz \
	openedge-function-manager/package.tar.gz \
	openedge-function-python27/package.tar.gz

SRC=$(wildcard *.go) $(shell find cmd master logger sdk protocol utils -type f -name '*.go')

openedge: $(SRC)
	@echo "BUILD $@"
	@go build ${GOFLAG} .

openedge-hub/package.tar.gz:
	make -C openedge-hub

openedge-agent/package.tar.gz:
	make -C openedge-agent

openedge-remote-mqtt/package.tar.gz:
	make -C openedge-remote-mqtt

openedge-function-manager/package.tar.gz:
	make -C openedge-function-manager

openedge-function-python27/package.tar.gz:
	make -C openedge-function-python27

test:
	go test --race ./...

tools: pubsub openedge-consistency

pubsub:
	@echo "BUILD $@"
	@go build ${GOFLAG} ./tools/pubsub

openedge-consistency:
	@echo "BUILD $@"
	@go build ${GOFLAG} ./tools/openedge-consistency

install: openedge
	install -d -m 0755 ${PREFIX}/bin
	install -m 0755 openedge ${PREFIX}/bin/
	tar cf - -C example/docker etc var | tar xvf - -C ${PREFIX}/

uninstall:
	rm -f ${PREFIX}/bin/openedge
	rm -rf ${PREFIX}/etc/openedge
	rm -rf ${PREFIX}/var/db/openedge
	rmdir ${PREFIX}/var/db
	rmdir ${PREFIX}/var
	rmdir ${PREFIX}/etc
	rmdir ${PREFIX}/bin
	rmdir ${PREFIX}

install-native: openedge package
	install -d -m 0755 ${PREFIX}/bin
	install -m 0755 openedge ${PREFIX}/bin/

	install -d -m 0755 ${PREFIX}/var/db/openedge/openedge-hub
	tar xzvf openedge-hub/package.tar.gz -C ${PREFIX}/var/db/openedge/openedge-hub

	install -d -m 0755 ${PREFIX}/var/db/openedge/openedge-agent
	tar xzvf openedge-agent/package.tar.gz -C ${PREFIX}/var/db/openedge/openedge-agent

	install -d -m 0755 ${PREFIX}/var/db/openedge/openedge-remote-mqtt
	tar xzvf openedge-remote-mqtt/package.tar.gz -C ${PREFIX}/var/db/openedge/openedge-remote-mqtt

	install -d -m 0755 ${PREFIX}/var/db/openedge/openedge-function-manager
	tar xzvf openedge-function-manager/package.tar.gz -C ${PREFIX}/var/db/openedge/openedge-function-manager

	install -d -m 0755 ${PREFIX}/var/db/openedge/openedge-function-python27
	tar xzvf openedge-function-python27/package.tar.gz -C ${PREFIX}/var/db/openedge/openedge-function-python27

	tar cf - -C example/native etc var | tar xvf - -C ${PREFIX}/

uninstall-native:
	rm -f ${PREFIX}/bin/openedge
	rm -rf ${PREFIX}/etc/openedge
	rm -rf ${PREFIX}/var/db/openedge
	rmdir ${PREFIX}/var/db
	rmdir ${PREFIX}/var
	rmdir ${PREFIX}/etc
	rmdir ${PREFIX}/bin
	rmdir ${PREFIX}

.PHONY: clean
clean:
	rm -f openedge
	make -C openedge-hub clean
	make -C openedge-agent clean
	make -C openedge-remote-mqtt clean
	make -C openedge-function-manager clean
	make -C openedge-function-python27 clean
	rm -f pubsub openedge-consistency

rebuild: clean all

generate:
	go generate ./...

image:
	make -C openedge-hub image
	make -C openedge-agent image
	make -C openedge-remote-mqtt image
	make -C openedge-function-manager image
	make -C openedge-function-python27 image

release:
	images-release
	# release linux 386
	env GOOS=linux GOARCH=386 make install PREFIX=__release_build/openedge-linux-386-$(VERSION)
	tar czf openedge-linux-386-$(VERSION).tar.gz -C __release_build/openedge-linux-386-$(VERSION) bin etc var
	tar cjf openedge-linux-386-$(VERSION).tar.bz2 -C __release_build/openedge-linux-386-$(VERSION) bin etc var
	make uninstall clean PREFIX=__release_build/openedge-linux-386-$(VERSION)
	# release linux amd64
	env GOOS=linux GOARCH=amd64 make install PREFIX=__release_build/openedge-linux-amd64-$(VERSION)
	tar czf openedge-linux-amd64-$(VERSION).tar.gz -C __release_build/openedge-linux-amd64-$(VERSION) bin etc var
	tar cjf openedge-linux-amd64-$(VERSION).tar.bz2 -C __release_build/openedge-linux-amd64-$(VERSION) bin etc var
	make uninstall clean PREFIX=__release_build/openedge-linux-amd64-$(VERSION)
	# release linux arm
	env GOOS=linux GOARCH=arm make install PREFIX=__release_build/openedge-linux-arm-$(VERSION)
	tar czf openedge-linux-arm-$(VERSION).tar.gz -C __release_build/openedge-linux-arm-$(VERSION) bin etc var
	tar cjf openedge-linux-arm-$(VERSION).tar.bz2 -C __release_build/openedge-linux-arm-$(VERSION) bin etc var
	make uninstall clean PREFIX=__release_build/openedge-linux-arm-$(VERSION)
	# release linux arm64
	env GOOS=linux GOARCH=arm64 make install PREFIX=__release_build/openedge-linux-arm64-$(VERSION)
	tar czf openedge-linux-arm64-$(VERSION).tar.gz -C __release_build/openedge-linux-arm64-$(VERSION) bin etc var
	tar cjf openedge-linux-arm64-$(VERSION).tar.bz2 -C __release_build/openedge-linux-arm64-$(VERSION) bin etc var
	make uninstall clean PREFIX=__release_build/openedge-linux-arm64-$(VERSION)
	# release darwin amd64
	env GOOS=darwin GOARCH=amd64 make all
	make install PREFIX=__release_build/openedge-darwin-amd64-$(VERSION)
	tar czf openedge-darwin-amd64-$(VERSION).tar.gz -C __release_build/openedge-darwin-amd64-$(VERSION) bin etc var
	tar cjf openedge-darwin-amd64-$(VERSION).tar.bz2 -C __release_build/openedge-darwin-amd64-$(VERSION) bin etc var
	make uninstall PREFIX=__release_build/openedge-darwin-amd64-$(VERSION)
	make install-native PREFIX=__release_build/openedge-darwin-amd64-$(VERSION)-native
	tar czf openedge-darwin-amd64-$(VERSION)-native.tar.gz -C __release_build/openedge-darwin-amd64-$(VERSION)-native bin etc var
	tar cjf openedge-darwin-amd64-$(VERSION)-native.tar.bz2 -C __release_build/openedge-darwin-amd64-$(VERSION)-native bin etc var
	make uninstall-native PREFIX=__release_build/openedge-darwin-amd64-$(VERSION)-native
	make clean
	# at last
	rmdir __release_build

images-release:
	# linux-amd64 images release
	env GOOS=linux GOARCH=amd64 make image IMAGE_PREFIX=$(REGISTRY) IMAGE_SUFFIX="-linux-amd64"
	make clean
	# linux-386 images release
	env GOOS=linux GOARCH=386 make image IMAGE_PREFIX=$(REGISTRY) IMAGE_SUFFIX="-linux-386"
	make clean
	# linux-arm images release
	env GOOS=linux GOARCH=arm make image IMAGE_PREFIX=$(REGISTRY) IMAGE_SUFFIX="-linux-arm"
	make clean
	# linux-arm64 images release
	env GOOS=linux GOARCH=arm64 make image IMAGE_PREFIX=$(REGISTRY) IMAGE_SUFFIX="-linux-arm64"
	make clean

# Need push built images first
manifest-push:
	mkdir tmp
	# Push openedge-agent manifest
	sed "s/__REGISTRY__/$(REGISTRY)/g; s/__NAMESPACE__/$(NAMESPACE)/g; s/__VERSION__/$(VERSION)/g;" openedge-agent/manifest.yml.template > tmp/manifest-agent.yml
	./bin/manifest-tool-linux-amd64 --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-agent.yml
	# Push openedge-hub manifest
	sed "s/__REGISTRY__/$(REGISTRY)/g; s/__NAMESPACE__/$(NAMESPACE)/g; s/__VERSION__/$(VERSION)/g;" openedge-hub/manifest.yml.template > tmp/manifest-hub.yml
	./bin/manifest-tool-linux-amd64 --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-hub.yml
	# Push openedge-function-manager manifest
	sed "s/__REGISTRY__/$(REGISTRY)/g; s/__NAMESPACE__/$(NAMESPACE)/g; s/__VERSION__/$(VERSION)/g;" openedge-function-manager/manifest.yml.template > tmp/manifest-function-manager.yml
	./bin/manifest-tool-linux-amd64 --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-function-manager.yml
	# Push openedge-function-python27 manifest
	sed "s/__REGISTRY__/$(REGISTRY)/g; s/__NAMESPACE__/$(NAMESPACE)/g; s/__VERSION__/$(VERSION)/g;" openedge-function-python27/manifest.yml.template > tmp/manifest-function-python27.yml
	./bin/manifest-tool-linux-amd64 --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-function-python27.yml
	# Push openedge-remote-mqtt manifest
	sed "s/__REGISTRY__/$(REGISTRY)/g; s/__NAMESPACE__/$(NAMESPACE)/g; s/__VERSION__/$(VERSION)/g;" openedge-remote-mqtt/manifest.yml.template > tmp/manifest-remote-mqtt.yml
	./bin/manifest-tool-linux-amd64 --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-remote-mqtt.yml

	rm -rf tmp
