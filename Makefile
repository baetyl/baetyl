PREFIX?=/usr/local
VERSION?=git-$(shell git rev-list HEAD|head -1|cut -c 1-6)
GOFLAG?=-ldflags "-X 'github.com/baidu/openedge/cmd.GoVersion=`go version`' -X 'github.com/baidu/openedge/cmd.Version=$(VERSION)'"
DEPLOY_TARGET=agent hub function-manager remote-mqtt timer function-python function-node

all: openedge package

package:
	for target in $(DEPLOY_TARGET) ; do \
		make openedge-$$target/package.zip;\
	done

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

openedge-function-python/package.zip:
	make -C openedge-function-python package27.zip
	make -C openedge-function-python package36.zip

openedge-function-node/package.zip:
	make -C openedge-function-node package85.zip

openedge-timer/package.zip:
	make -C openedge-timer

test:
	go test --race --cover ./...

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
	rm -rf ${PREFIX}/var/log/openedge
	rm -rf ${PREFIX}/var/run/openedge
	rmdir ${PREFIX}/var/db
	rmdir ${PREFIX}/var/log
	rmdir ${PREFIX}/var/run
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

image: clean
	for target in $(DEPLOY_TARGET) ; do \
		make -C openedge-$$target image;\
	done

function-python-image:
	make -C openedge-function-python image

release: clean release-master release-image push-image release-manifest release-package

release-master: clean
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

release-image: clean
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

release-manifest:
	rm -rf tmp
	mkdir tmp
	for target in $(DEPLOY_TARGET) ; do \
		sed "s/__REGISTRY__/$(REGISTRY)/g; s/__NAMESPACE__/$(NAMESPACE)/g; s/__VERSION__/$(VERSION)/g;" openedge-$$target/manifest.yml.template > tmp/manifest-$$target-$(VERSION).yml;\
		./bin/manifest-tool-linux-amd64 --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-$$target-$(VERSION).yml;\
		sed "s/__REGISTRY__/$(REGISTRY)/g; s/__NAMESPACE__/$(NAMESPACE)/g; s/__VERSION__/latest/g;" openedge-$$target/manifest.yml.template > tmp/manifest-$$target-latest.yml;\
		./bin/manifest-tool-linux-amd64 --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-$$target-latest.yml;\
	done
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

release-package: clean
	# Release modules' package -- linux armv7
	env GOOS=linux GOARCH=arm GOARM=7 make package
	for target in $(DEPLOY_TARGET) ; do \
		mv openedge-$$target/package.zip ./openedge-$$target-linux-armv7-$(VERSION).zip;\
	done
	make clean
	# Release modules' package -- linux amd64
	env GOOS=linux GOARCH=amd64 make package
	for target in $(DEPLOY_TARGET); do \
		mv openedge-$$target/package.zip ./openedge-$$target-linux-amd64-$(VERSION).zip;\
	done
	make clean
	# Release modules' package -- linux arm64
	env GOOS=linux GOARCH=arm64 make package
	for target in $(DEPLOY_TARGET); do \
		mv openedge-$$target/package.zip ./openedge-$$target-linux-arm64-$(VERSION).zip;\
	done
	make clean
	# Release modules' package -- linux 386
	env GOOS=linux GOARCH=386 make package
	for target in $(DEPLOY_TARGET); do \
		mv openedge-$$target/package.zip ./openedge-$$target-linux-386-$(VERSION).zip;\
	done
	make clean
	# Release modules' package -- darwin amd64
	env GOOS=darwin GOARCH=amd64 make package
	for target in $(DEPLOY_TARGET); do \
		mv openedge-$$target/package.zip ./openedge-$$target-darwin-amd64-$(VERSION).zip;\
	done
	make clean

push-image:
	for target in $(DEPLOY_TARGET); do \
		docker tag $(IMAGE_PREFIX)openedge-$$target-linux-amd64:latest $(IMAGE_PREFIX)openedge-$$target-linux-amd64:$(VERSION);\
		docker tag $(IMAGE_PREFIX)openedge-$$target-linux-arm64:latest $(IMAGE_PREFIX)openedge-$$target-linux-arm64:$(VERSION);\
		docker tag $(IMAGE_PREFIX)openedge-$$target-linux-armv7:latest $(IMAGE_PREFIX)openedge-$$target-linux-armv7:$(VERSION);\
		docker tag $(IMAGE_PREFIX)openedge-$$target-linux-386:latest $(IMAGE_PREFIX)openedge-$$target-linux-386:$(VERSION);\
		docker push $(IMAGE_PREFIX)openedge-$$target-linux-amd64;\
		docker push $(IMAGE_PREFIX)openedge-$$target-linux-arm64;\
		docker push $(IMAGE_PREFIX)openedge-$$target-linux-armv7;\
		docker push $(IMAGE_PREFIX)openedge-$$target-linux-386;\
	done