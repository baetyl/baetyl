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
	@go build -o openedge ${GOFLAG} .

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

.PHONY: uninstall
uninstall:
	@rm -f ${PREFIX}/bin/openedge
	@rm -rf ${PREFIX}/etc/openedge
	@rm -rf ${PREFIX}/var/db/openedge
	@rm -rf ${PREFIX}/var/log/openedge
	@rm -rf ${PREFIX}/var/run/openedge
	@! test -d ${PREFIX}/var/db || ! ls ${PREFIX}/var/db | xargs test -z || rmdir ${PREFIX}/var/db
	@! test -d ${PREFIX}/var/log || ! ls ${PREFIX}/var/log | xargs test -z || rmdir ${PREFIX}/var/log
	@! test -d ${PREFIX}/var/run || ! ls ${PREFIX}/var/run | xargs test -z || rmdir ${PREFIX}/var/run
	@! test -d ${PREFIX}/var || ! ls ${PREFIX}/var | xargs test -z || rmdir ${PREFIX}/var
	@! test -d ${PREFIX}/etc || ! ls ${PREFIX}/etc | xargs test -z || rmdir ${PREFIX}/etc
	@! test -d ${PREFIX}/bin || ! ls ${PREFIX}/bin | xargs test -z || rmdir ${PREFIX}/bin
	@! test -d ${PREFIX} || ! ls ${PREFIX} | xargs test -z || rmdir ${PREFIX}

.PHONY: uninstall-native
uninstall-native: uninstall

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

release: clean release-master release-image push-image release-manifest

release-master: clean
	# release linux 386
	env GOOS=linux GOARCH=386 make install PREFIX=__release_build/openedge-$(VERSION)-linux-386
	cd __release_build/openedge-$(VERSION)-linux-386 && zip -q -r ../../openedge-$(VERSION)-linux-386.zip bin/
	make uninstall clean PREFIX=__release_build/openedge-$(VERSION)-linux-386
	# release linux amd64
	env GOOS=linux GOARCH=amd64 make install PREFIX=__release_build/openedge-$(VERSION)-linux-amd64
	cd __release_build/openedge-$(VERSION)-linux-amd64 && zip -q -r ../../openedge-$(VERSION)-linux-amd64.zip bin/
	make uninstall clean PREFIX=__release_build/openedge-$(VERSION)-linux-amd64
	# release linux arm v7
	env GOOS=linux GOARCH=arm GOARM=7 make install PREFIX=__release_build/openedge-$(VERSION)-linux-armv7
	cd __release_build/openedge-$(VERSION)-linux-armv7 && zip -q -r ../../openedge-$(VERSION)-linux-armv7.zip bin/
	make uninstall clean PREFIX=__release_build/openedge-$(VERSION)-linux-armv7
	# release linux arm64
	env GOOS=linux GOARCH=arm64 make install PREFIX=__release_build/openedge-$(VERSION)-linux-arm64
	cd __release_build/openedge-$(VERSION)-linux-arm64 && zip -q -r ../../openedge-$(VERSION)-linux-arm64.zip bin/
	make uninstall clean PREFIX=__release_build/openedge-$(VERSION)-linux-arm64
	# release darwin amd64
	env GOOS=darwin GOARCH=amd64 make install PREFIX=__release_build/openedge-$(VERSION)-darwin-amd64
	cd __release_build/openedge-$(VERSION)-darwin-amd64 && zip -q -r ../../openedge-$(VERSION)-darwin-amd64.zip bin/
	make uninstall PREFIX=__release_build/openedge-$(VERSION)-darwin-amd64
	make clean
	# at last
	rmdir __release_build

release-image: clean
	# linux-amd64 images release
	env GOOS=linux GOARCH=amd64 make image TAG="$(VERSION)-linux-amd64"
	make clean
	# linux-386 images release
	env GOOS=linux GOARCH=386 make image TAG="$(VERSION)-linux-386"
	make clean
	# linux-arm images release
	env GOOS=linux GOARCH=arm GOARM=7 make image TAG="$(VERSION)-linux-armv7"
	make clean
	# linux-arm64 images release
	env GOOS=linux GOARCH=arm64 make image TAG="$(VERSION)-linux-arm64"
	make clean

release-manifest:
	rm -rf tmp
	mkdir tmp
	for target in $(DEPLOY_TARGET) ; do \
		sed "s?__IMAGE_PREFIX__/?$(IMAGE_PREFIX)?g; s?__TAG__?$(VERSION)?g; s?__VERSION__?$(VERSION)?g; s?__TARGET__?openedge-$$target?g;" manifest.yml.template > tmp/manifest-$$target-$(VERSION).yml;\
		./bin/manifest-tool-linux-amd64 --insecure --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-$$target-$(VERSION).yml;\
	done
	rm -rf tmp

release-manifest-latest:
	rm -rf tmp
	mkdir tmp
	for target in $(DEPLOY_TARGET) ; do \
		sed "s?__IMAGE_PREFIX__/?$(IMAGE_PREFIX)?g; s?__TAG__?latest?g; s?__VERSION__?$(VERSION)?g; s?__TARGET__?openedge-$$target?g;" manifest.yml.template > tmp/manifest-$$target-latest.yml;\
		./bin/manifest-tool-linux-amd64 --insecure --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-$$target-$(VERSION).yml;\
	done
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
		docker push $(IMAGE_PREFIX)openedge-$$target:$(VERSION)-linux-amd64;\
		docker push $(IMAGE_PREFIX)openedge-$$target:$(VERSION)-linux-arm64;\
		docker push $(IMAGE_PREFIX)openedge-$$target:$(VERSION)-linux-armv7;\
		docker push $(IMAGE_PREFIX)openedge-$$target:$(VERSION)-linux-386;\
	done
