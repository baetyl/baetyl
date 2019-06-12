PREFIX?=/usr/local
VERSION?=git-$(shell git rev-list HEAD|head -1|cut -c 1-6)
GOFLAG?=-ldflags "-X 'github.com/baidu/openedge/cmd.GoVersion=`go version`' -X 'github.com/baidu/openedge/cmd.Version=$(VERSION)'"

all: openedge package

package: \
	openedge-hub/package.zip \
	openedge-agent/package.zip \
	openedge-remote-mqtt/package.zip \
	openedge-function-manager/package.zip \
	openedge-function-python/package27.zip \
	openedge-function-python/package36.zip \
	openedge-function-node/package85.zip \
	openedge-timer/package.zip

SRC=$(wildcard *.go) $(shell find cmd master logger sdk protocol utils -type f -name '*.go')

openedge: $(SRC)
	@echo "BUILD $@"
	@go build ${GOFLAG} .

openedge-hub/package.zip:
	make -C openedge-hub

openedge-agent/package.zip:
	make -C openedge-agent

openedge-remote-mqtt/package.zip:
	make -C openedge-remote-mqtt

openedge-function-manager/package.zip:
	make -C openedge-function-manager

openedge-function-python/package27.zip:
	make -C openedge-function-python package27.zip

openedge-function-python/package36.zip:
	make -C openedge-function-python package36.zip

openedge-function-node/package85.zip:
	make -C openedge-function-node package85.zip

openedge-timer/package.zip:
	make -C openedge-timer

test:
	go test --race ./...

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
	unzip -o openedge-hub/package.zip -d ${PREFIX}/var/db/openedge/openedge-hub

	install -d -m 0755 ${PREFIX}/var/db/openedge/openedge-agent
	unzip -o openedge-agent/package.zip -d ${PREFIX}/var/db/openedge/openedge-agent

	install -d -m 0755 ${PREFIX}/var/db/openedge/openedge-remote-mqtt
	unzip -o openedge-remote-mqtt/package.zip -d ${PREFIX}/var/db/openedge/openedge-remote-mqtt

	install -d -m 0755 ${PREFIX}/var/db/openedge/openedge-function-manager
	unzip -o openedge-function-manager/package.zip -d ${PREFIX}/var/db/openedge/openedge-function-manager

	install -d -m 0755 ${PREFIX}/var/db/openedge/openedge-function-python27
	unzip -o openedge-function-python/package27.zip -d ${PREFIX}/var/db/openedge/openedge-function-python27

	install -d -m 0755 ${PREFIX}/var/db/openedge/openedge-function-python36
	unzip -o openedge-function-python/package36.zip -d ${PREFIX}/var/db/openedge/openedge-function-python36

	install -d -m 0755 ${PREFIX}/var/db/openedge/openedge-function-node85
	unzip -o openedge-function-node/package85.zip -d ${PREFIX}/var/db/openedge/openedge-function-node85

	install -d -m 0755 ${PREFIX}/var/db/openedge/openedge-timer
	unzip -o openedge-timer/package.zip -d ${PREFIX}/var/db/openedge/openedge-timer

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
	make -C openedge-function-python clean
	make -C openedge-function-node clean
	make -C openedge-timer clean

rebuild: clean all

generate:
	go generate ./...

image:
	make -C openedge-hub image
	make -C openedge-agent image
	make -C openedge-remote-mqtt image
	make -C openedge-function-manager image
	make -C openedge-timer image
	make function-python-image
	make -C openedge-function-node image

function-python-image:
	make -C openedge-function-python image

release: release-image push-image release-manifest release-package
	# release linux 386
	env GOOS=linux GOARCH=386 make install PREFIX=__release_build/openedge-linux-386-$(VERSION)
	tar czf openedge-linux-386-$(VERSION).tar.gz -C __release_build/openedge-linux-386-$(VERSION) bin etc var
	tar cjf openedge-linux-386-$(VERSION).tar.bz2 -C __release_build/openedge-linux-386-$(VERSION) bin etc var
	cd __release_build/openedge-linux-386-$(VERSION) && zip -q -r ../../openedge-linux-386-$(VERSION).zip bin/
	make uninstall clean PREFIX=__release_build/openedge-linux-386-$(VERSION)
	# release linux amd64
	env GOOS=linux GOARCH=amd64 make install PREFIX=__release_build/openedge-linux-amd64-$(VERSION)
	tar czf openedge-linux-amd64-$(VERSION).tar.gz -C __release_build/openedge-linux-amd64-$(VERSION) bin etc var
	tar cjf openedge-linux-amd64-$(VERSION).tar.bz2 -C __release_build/openedge-linux-amd64-$(VERSION) bin etc var
	cd __release_build/openedge-linux-amd64-$(VERSION) && zip -q -r ../../openedge-linux-amd64-$(VERSION).zip bin/
	make uninstall clean PREFIX=__release_build/openedge-linux-amd64-$(VERSION)
	# release linux arm v7
	env GOOS=linux GOARCH=arm GOARM=7 make install PREFIX=__release_build/openedge-linux-armv7-$(VERSION)
	tar czf openedge-linux-armv7-$(VERSION).tar.gz -C __release_build/openedge-linux-armv7-$(VERSION) bin etc var
	tar cjf openedge-linux-armv7-$(VERSION).tar.bz2 -C __release_build/openedge-linux-armv7-$(VERSION) bin etc var
	cd __release_build/openedge-linux-armv7-$(VERSION) && zip -q -r ../../openedge-linux-armv7-$(VERSION).zip bin/
	make uninstall clean PREFIX=__release_build/openedge-linux-armv7-$(VERSION)
	# release linux arm64
	env GOOS=linux GOARCH=arm64 make install PREFIX=__release_build/openedge-linux-arm64-$(VERSION)
	tar czf openedge-linux-arm64-$(VERSION).tar.gz -C __release_build/openedge-linux-arm64-$(VERSION) bin etc var
	tar cjf openedge-linux-arm64-$(VERSION).tar.bz2 -C __release_build/openedge-linux-arm64-$(VERSION) bin etc var
	cd __release_build/openedge-linux-arm64-$(VERSION) && zip -q -r ../../openedge-linux-arm64-$(VERSION).zip bin/
	make uninstall clean PREFIX=__release_build/openedge-linux-arm64-$(VERSION)
	# release darwin amd64
	env GOOS=darwin GOARCH=amd64 make all
	make install PREFIX=__release_build/openedge-darwin-amd64-$(VERSION)
	tar czf openedge-darwin-amd64-$(VERSION).tar.gz -C __release_build/openedge-darwin-amd64-$(VERSION) bin etc var
	tar cjf openedge-darwin-amd64-$(VERSION).tar.bz2 -C __release_build/openedge-darwin-amd64-$(VERSION) bin etc var
	cd __release_build/openedge-darwin-amd64-$(VERSION) && zip -q -r ../../openedge-darwin-amd64-$(VERSION).zip bin/
	make uninstall PREFIX=__release_build/openedge-darwin-amd64-$(VERSION)
	make install-native PREFIX=__release_build/openedge-darwin-amd64-$(VERSION)-native
	tar czf openedge-darwin-amd64-$(VERSION)-native.tar.gz -C __release_build/openedge-darwin-amd64-$(VERSION)-native bin etc var
	tar cjf openedge-darwin-amd64-$(VERSION)-native.tar.bz2 -C __release_build/openedge-darwin-amd64-$(VERSION)-native bin etc var
	make uninstall-native PREFIX=__release_build/openedge-darwin-amd64-$(VERSION)-native
	make clean
	# at last
	rmdir __release_build

release-image:
	# linux-amd64 images release
	env GOOS=linux GOARCH=amd64 make image IMAGE_SUFFIX="-linux-amd64"
	make clean
	# linux-386 images release
	env GOOS=linux GOARCH=386 make image IMAGE_SUFFIX="-linux-386"
	make clean
	# linux-arm images release
	env GOOS=linux GOARCH=arm GOARM=7 make image IMAGE_SUFFIX="-linux-armv7"
	make clean
	# linux-arm64 images release
	env GOOS=linux GOARCH=arm64 make image IMAGE_SUFFIX="-linux-arm64"
	make clean

# Need push built images first
release-manifest:
	rm -rf tmp
	mkdir tmp
	# Push openedge-agent manifest version
	sed "s/__REGISTRY__/$(REGISTRY)/g; s/__NAMESPACE__/$(NAMESPACE)/g; s/__VERSION__/$(VERSION)/g;" openedge-agent/manifest.yml.template > tmp/manifest-agent-$(VERSION).yml
	./bin/manifest-tool-linux-amd64 --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-agent-$(VERSION).yml
	# Push openedge-agent manifest latest
	sed "s/__REGISTRY__/$(REGISTRY)/g; s/__NAMESPACE__/$(NAMESPACE)/g; s/__VERSION__/latest/g;" openedge-agent/manifest.yml.template > tmp/manifest-agent-latest.yml
	./bin/manifest-tool-linux-amd64 --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-agent-latest.yml
	# Push openedge-hub manifest version
	sed "s/__REGISTRY__/$(REGISTRY)/g; s/__NAMESPACE__/$(NAMESPACE)/g; s/__VERSION__/$(VERSION)/g;" openedge-hub/manifest.yml.template > tmp/manifest-hub-$(VERSION).yml
	./bin/manifest-tool-linux-amd64 --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-hub-$(VERSION).yml
	# Push openedge-hub manifest latest
	sed "s/__REGISTRY__/$(REGISTRY)/g; s/__NAMESPACE__/$(NAMESPACE)/g; s/__VERSION__/latest/g;" openedge-hub/manifest.yml.template > tmp/manifest-hub-latest.yml
	./bin/manifest-tool-linux-amd64 --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-hub-latest.yml
	# Push openedge-function-manager manifest version
	sed "s/__REGISTRY__/$(REGISTRY)/g; s/__NAMESPACE__/$(NAMESPACE)/g; s/__VERSION__/$(VERSION)/g;" openedge-function-manager/manifest.yml.template > tmp/manifest-function-manager-$(VERSION).yml
	./bin/manifest-tool-linux-amd64 --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-function-manager-$(VERSION).yml
	# Push openedge-function-manager manifest latest
	sed "s/__REGISTRY__/$(REGISTRY)/g; s/__NAMESPACE__/$(NAMESPACE)/g; s/__VERSION__/latest/g;" openedge-function-manager/manifest.yml.template > tmp/manifest-function-manager-latest.yml
	./bin/manifest-tool-linux-amd64 --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-function-manager-latest.yml
	# Push openedge-remote-mqtt manifest version
	sed "s/__REGISTRY__/$(REGISTRY)/g; s/__NAMESPACE__/$(NAMESPACE)/g; s/__VERSION__/$(VERSION)/g;" openedge-remote-mqtt/manifest.yml.template > tmp/manifest-remote-mqtt-$(VERSION).yml
	./bin/manifest-tool-linux-amd64 --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-remote-mqtt-$(VERSION).yml
	# Push openedge-remote-mqtt manifest latest
	sed "s/__REGISTRY__/$(REGISTRY)/g; s/__NAMESPACE__/$(NAMESPACE)/g; s/__VERSION__/latest/g;" openedge-remote-mqtt/manifest.yml.template > tmp/manifest-remote-mqtt-latest.yml
	./bin/manifest-tool-linux-amd64 --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-remote-mqtt-latest.yml
	# Push openedge-timer manifest version
	sed "s/__REGISTRY__/$(REGISTRY)/g; s/__NAMESPACE__/$(NAMESPACE)/g; s/__VERSION__/$(VERSION)/g;" openedge-timer/manifest.yml.template > tmp/manifest-timer-$(VERSION).yml
	./bin/manifest-tool-linux-amd64 --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-timer-$(VERSION).yml
	# Push openedge-timer manifest latest
	sed "s/__REGISTRY__/$(REGISTRY)/g; s/__NAMESPACE__/$(NAMESPACE)/g; s/__VERSION__/latest/g;" openedge-timer/manifest.yml.template > tmp/manifest-timer-latest.yml
	./bin/manifest-tool-linux-amd64 --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-timer-latest.yml

	rm -rf tmp

# You need build the function-builder image at different platforms and push them to the hub first
release-builder-manifest:
	mkdir tmp

	# Push openedge-python27-builder manifest version
	sed "s/__REGISTRY__/$(REGISTRY)/g; s/__NAMESPACE__/$(NAMESPACE)/g; s/__VERSION__/$(VERSION)/g;" openedge-function-python/manifest-python27-builder.yml.template > tmp/manifest-python27-builder-$(VERSION).yml
	./bin/manifest-tool-linux-amd64 --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-python27-builder-$(VERSION).yml
	# Push openedge-python27-builder manifest latest
	sed "s/__REGISTRY__/$(REGISTRY)/g; s/__NAMESPACE__/$(NAMESPACE)/g; s/__VERSION__/latest/g;" openedge-function-python/manifest-python27-builder.yml.template > tmp/manifest-python27-builder-latest.yml
	./bin/manifest-tool-linux-amd64 --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-python27-builder-latest.yml
	# Push openedge-python36-builder manifest version
	sed "s/__REGISTRY__/$(REGISTRY)/g; s/__NAMESPACE__/$(NAMESPACE)/g; s/__VERSION__/$(VERSION)/g;" openedge-function-python/manifest-python36-builder.yml.template > tmp/manifest-python36-builder-$(VERSION).yml
	./bin/manifest-tool-linux-amd64 --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-python36-builder-$(VERSION).yml
	# Push openedge-python36-builder manifest latest
	sed "s/__REGISTRY__/$(REGISTRY)/g; s/__NAMESPACE__/$(NAMESPACE)/g; s/__VERSION__/latest/g;" openedge-function-python/manifest-python36-builder.yml.template > tmp/manifest-python36-builder-latest.yml
	./bin/manifest-tool-linux-amd64 --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-python36-builder-latest.yml
	# Push openedge-node85-builder manifest version
	sed "s/__REGISTRY__/$(REGISTRY)/g; s/__NAMESPACE__/$(NAMESPACE)/g; s/__VERSION__/$(VERSION)/g;" openedge-function-node/manifest-node85-builder.yml.template > tmp/manifest-node85-builder-$(VERSION).yml
	./bin/manifest-tool-linux-amd64 --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-node85-builder-$(VERSION).yml
	# Push openedge-node85-builder manifest latest
	sed "s/__REGISTRY__/$(REGISTRY)/g; s/__NAMESPACE__/$(NAMESPACE)/g; s/__VERSION__/latest/g;" openedge-function-node/manifest-node85-builder.yml.template > tmp/manifest-node85-builder-latest.yml
	./bin/manifest-tool-linux-amd64 --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-node85-builder-latest.yml

	rm -rf tmp

release-package:
	# Release modules' package -- linux arm
	env GOOS=linux GOARCH=arm GOARM=7 make package
	mv openedge-agent/package.zip ./openedge-agent-linux-armv7-$(VERSION).zip
	mv openedge-hub/package.zip ./openedge-hub-linux-armv7-$(VERSION).zip
	mv openedge-remote-mqtt/package.zip ./openedge-remote-mqtt-linux-armv7-$(VERSION).zip
	mv openedge-function-manager/package.zip ./openedge-function-manager-linux-armv7-$(VERSION).zip
	mv openedge-function-python/package27.zip ./openedge-function-python27-linux-armv7-$(VERSION).zip
	mv openedge-function-python/package36.zip ./openedge-function-python36-linux-armv7-$(VERSION).zip
	mv openedge-function-node/package85.zip ./openedge-function-node85-linux-armv7-$(VERSION).zip
	mv openedge-timer/package.zip ./openedge-timer-linux-armv7-$(VERSION).zip
	make clean
	# Release modules' package -- linux amd64
	env GOOS=linux GOARCH=amd64 make package
	mv openedge-agent/package.zip ./openedge-agent-linux-amd64-$(VERSION).zip
	mv openedge-hub/package.zip ./openedge-hub-linux-amd64-$(VERSION).zip
	mv openedge-remote-mqtt/package.zip ./openedge-remote-mqtt-linux-amd64-$(VERSION).zip
	mv openedge-function-manager/package.zip ./openedge-function-manager-linux-amd64-$(VERSION).zip
	mv openedge-function-python/package27.zip ./openedge-function-python27-linux-amd64-$(VERSION).zip
	mv openedge-function-python/package36.zip ./openedge-function-python36-linux-amd64-$(VERSION).zip
	mv openedge-function-node/package85.zip ./openedge-function-node85-linux-amd64-$(VERSION).zip
	mv openedge-timer/package.zip ./openedge-timer-linux-amd64-$(VERSION).zip
	make clean
	# Release modules' package -- linux arm64
	env GOOS=linux GOARCH=arm64 make package
	mv openedge-agent/package.zip ./openedge-agent-linux-arm64-$(VERSION).zip
	mv openedge-hub/package.zip ./openedge-hub-linux-arm64-$(VERSION).zip
	mv openedge-remote-mqtt/package.zip ./openedge-remote-mqtt-linux-arm64-$(VERSION).zip
	mv openedge-function-manager/package.zip ./openedge-function-manager-linux-arm64-$(VERSION).zip
	mv openedge-function-python/package27.zip ./openedge-function-python27-linux-arm64-$(VERSION).zip
	mv openedge-function-python/package36.zip ./openedge-function-python36-linux-arm64-$(VERSION).zip
	mv openedge-function-node/package85.zip ./openedge-function-node85-linux-arm64-$(VERSION).zip
	mv openedge-timer/package.zip ./openedge-timer-linux-arm64-$(VERSION).zip
	make clean
	# Release modules' package -- linux 386
	env GOOS=linux GOARCH=386 make package
	mv openedge-agent/package.zip ./openedge-agent-linux-386-$(VERSION).zip
	mv openedge-hub/package.zip ./openedge-hub-linux-386-$(VERSION).zip
	mv openedge-remote-mqtt/package.zip ./openedge-remote-mqtt-linux-386-$(VERSION).zip
	mv openedge-function-manager/package.zip ./openedge-function-manager-linux-386-$(VERSION).zip
	mv openedge-function-python/package27.zip ./openedge-function-python27-linux-386-$(VERSION).zip
	mv openedge-function-python/package36.zip ./openedge-function-python36-linux-386-$(VERSION).zip
	mv openedge-function-node/package85.zip ./openedge-function-node85-linux-386-$(VERSION).zip
	mv openedge-timer/package.zip ./openedge-timer-linux-386-$(VERSION).zip
	make clean
	# Release modules' package -- darwin amd64
	env GOOS=darwin GOARCH=amd64 make package
	mv openedge-agent/package.zip ./openedge-agent-darwin-amd64-$(VERSION).zip
	mv openedge-hub/package.zip ./openedge-hub-darwin-amd64-$(VERSION).zip
	mv openedge-remote-mqtt/package.zip ./openedge-remote-mqtt-darwin-amd64-$(VERSION).zip
	mv openedge-function-manager/package.zip ./openedge-function-manager-darwin-amd64-$(VERSION).zip
	mv openedge-function-python/package27.zip ./openedge-function-python27-darwin-amd64-$(VERSION).zip
	mv openedge-function-python/package36.zip ./openedge-function-python36-darwin-amd64-$(VERSION).zip
	mv openedge-function-node/package85.zip ./openedge-function-node85-darwin-amd64-$(VERSION).zip
	mv openedge-timer/package.zip ./openedge-timer-darwin-amd64-$(VERSION).zip
	make clean

push-image:
	# Push hub images
	docker tag $(IMAGE_PREFIX)openedge-hub-linux-amd64:latest $(IMAGE_PREFIX)openedge-hub-linux-amd64:$(VERSION)
	docker tag $(IMAGE_PREFIX)openedge-hub-linux-arm64:latest $(IMAGE_PREFIX)openedge-hub-linux-arm64:$(VERSION)
	docker tag $(IMAGE_PREFIX)openedge-hub-linux-arm:latest $(IMAGE_PREFIX)openedge-hub-linux-arm:$(VERSION)
	docker tag $(IMAGE_PREFIX)openedge-hub-linux-386:latest $(IMAGE_PREFIX)openedge-hub-linux-386:$(VERSION)
	docker push $(IMAGE_PREFIX)openedge-hub-linux-amd64
	docker push $(IMAGE_PREFIX)openedge-hub-linux-arm64
	docker push $(IMAGE_PREFIX)openedge-hub-linux-arm
	docker push $(IMAGE_PREFIX)openedge-hub-linux-386
	# Push agent images
	docker tag $(IMAGE_PREFIX)openedge-agent-linux-amd64:latest $(IMAGE_PREFIX)openedge-agent-linux-amd64:$(VERSION)
	docker tag $(IMAGE_PREFIX)openedge-agent-linux-arm64:latest $(IMAGE_PREFIX)openedge-agent-linux-arm64:$(VERSION)
	docker tag $(IMAGE_PREFIX)openedge-agent-linux-arm:latest $(IMAGE_PREFIX)openedge-agent-linux-arm:$(VERSION)
	docker tag $(IMAGE_PREFIX)openedge-agent-linux-386:latest $(IMAGE_PREFIX)openedge-agent-linux-386:$(VERSION)
	docker push $(IMAGE_PREFIX)openedge-agent-linux-amd64
	docker push $(IMAGE_PREFIX)openedge-agent-linux-arm64
	docker push $(IMAGE_PREFIX)openedge-agent-linux-arm
	docker push $(IMAGE_PREFIX)openedge-agent-linux-386
	# Push function manager images
	docker tag $(IMAGE_PREFIX)openedge-function-manager-linux-amd64:latest $(IMAGE_PREFIX)openedge-function-manager-linux-amd64:$(VERSION)
	docker tag $(IMAGE_PREFIX)openedge-function-manager-linux-arm64:latest $(IMAGE_PREFIX)openedge-function-manager-linux-arm64:$(VERSION)
	docker tag $(IMAGE_PREFIX)openedge-function-manager-linux-arm:latest $(IMAGE_PREFIX)openedge-function-manager-linux-arm:$(VERSION)
	docker tag $(IMAGE_PREFIX)openedge-function-manager-linux-386:latest $(IMAGE_PREFIX)openedge-function-manager-linux-386:$(VERSION)
	docker push $(IMAGE_PREFIX)openedge-function-manager-linux-amd64
	docker push $(IMAGE_PREFIX)openedge-function-manager-linux-arm64
	docker push $(IMAGE_PREFIX)openedge-function-manager-linux-arm
	docker push $(IMAGE_PREFIX)openedge-function-manager-linux-386
	# Push remote mqtt images
	docker tag $(IMAGE_PREFIX)openedge-remote-mqtt-linux-amd64:latest $(IMAGE_PREFIX)openedge-remote-mqtt-linux-amd64:$(VERSION)
	docker tag $(IMAGE_PREFIX)openedge-remote-mqtt-linux-arm64:latest $(IMAGE_PREFIX)openedge-remote-mqtt-linux-arm64:$(VERSION)
	docker tag $(IMAGE_PREFIX)openedge-remote-mqtt-linux-arm:latest $(IMAGE_PREFIX)openedge-remote-mqtt-linux-arm:$(VERSION)
	docker tag $(IMAGE_PREFIX)openedge-remote-mqtt-linux-386:latest $(IMAGE_PREFIX)openedge-remote-mqtt-linux-386:$(VERSION)
	docker push $(IMAGE_PREFIX)openedge-remote-mqtt-linux-amd64
	docker push $(IMAGE_PREFIX)openedge-remote-mqtt-linux-arm64
	docker push $(IMAGE_PREFIX)openedge-remote-mqtt-linux-arm
	docker push $(IMAGE_PREFIX)openedge-remote-mqtt-linux-386
	# Push timer images
	docker tag $(IMAGE_PREFIX)openedge-timer-linux-amd64:latest $(IMAGE_PREFIX)openedge-timer-linux-amd64:$(VERSION)
	docker tag $(IMAGE_PREFIX)openedge-timer-linux-arm64:latest $(IMAGE_PREFIX)openedge-timer-linux-arm64:$(VERSION)
	docker tag $(IMAGE_PREFIX)openedge-timer-linux-arm:latest $(IMAGE_PREFIX)openedge-timer-linux-arm:$(VERSION)
	docker tag $(IMAGE_PREFIX)openedge-timer-linux-386:latest $(IMAGE_PREFIX)openedge-timer-linux-386:$(VERSION)
	docker push $(IMAGE_PREFIX)openedge-timer-linux-amd64
	docker push $(IMAGE_PREFIX)openedge-timer-linux-arm64
	docker push $(IMAGE_PREFIX)openedge-timer-linux-arm
	docker push $(IMAGE_PREFIX)openedge-timer-linux-386
	# Push function python27 images
	docker tag $(IMAGE_PREFIX)openedge-function-python27:latest $(IMAGE_PREFIX)openedge-function-python27:$(VERSION)
	docker push $(IMAGE_PREFIX)openedge-function-python27
	# Push function python36 images
	docker tag $(IMAGE_PREFIX)openedge-function-python36:latest $(IMAGE_PREFIX)openedge-function-python36:$(VERSION)
	docker push $(IMAGE_PREFIX)openedge-function-python36
	# Push function node85 images
	docker tag $(IMAGE_PREFIX)openedge-function-node85:latest $(IMAGE_PREFIX)openedge-function-node85:$(VERSION)
	docker push $(IMAGE_PREFIX)openedge-function-node85
