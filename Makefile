PREFIX?=/usr/local
REVISION?=git-$(shell git rev-list HEAD|head -1|cut -c 1-6)
VERSION?=$(REVISION)
GOFLAG?=-ldflags "-X 'github.com/baetyl/baetyl/cmd.Revision=$(REVISION)' -X 'github.com/baetyl/baetyl/cmd.Version=$(VERSION)'"
GOTESTFLAG?=
GOTESTPKGS?=$(shell go list ./... | grep -v baetyl-video-infer)
DEPLOY_TARGET=agent hub function-manager remote-mqtt timer function-python function-node

all: baetyl package

package:
	for target in $(DEPLOY_TARGET) ; do \
		make baetyl-$$target/package.zip;\
	done

SRC=$(wildcard *.go) $(shell find cmd master logger sdk protocol utils -type f -name '*.go')

baetyl: $(SRC)
	@echo "BUILD $@"
	@go build -o baetyl ${GOFLAG} .

baetyl-hub/package.zip:
	make -C baetyl-hub

baetyl-agent/package.zip:
	make -C baetyl-agent

baetyl-remote-mqtt/package.zip:
	make -C baetyl-remote-mqtt

baetyl-function-manager/package.zip:
	make -C baetyl-function-manager

baetyl-function-python/package.zip:
	make -C baetyl-function-python package27.zip
	make -C baetyl-function-python package36.zip

baetyl-function-node/package.zip:
	make -C baetyl-function-node package85.zip

baetyl-timer/package.zip:
	make -C baetyl-timer

test:
	go test ${GOTESTFLAG} -coverprofile=coverage.out ${GOTESTPKGS}
	go tool cover -func=coverage.out | grep total

install: baetyl
	install -d -m 0755 ${PREFIX}/bin
	install -m 0755 baetyl ${PREFIX}/bin/
	tar cf - -C example/docker etc var | tar xvf - -C ${PREFIX}/

install-native: baetyl package
	install -d -m 0755 ${PREFIX}/bin
	install -m 0755 baetyl ${PREFIX}/bin/

	install -d -m 0755 ${PREFIX}/var/db/baetyl/baetyl-hub
	unzip -o baetyl-hub/package.zip -d ${PREFIX}/var/db/baetyl/baetyl-hub

	install -d -m 0755 ${PREFIX}/var/db/baetyl/baetyl-agent
	unzip -o baetyl-agent/package.zip -d ${PREFIX}/var/db/baetyl/baetyl-agent

	install -d -m 0755 ${PREFIX}/var/db/baetyl/baetyl-remote-mqtt
	unzip -o baetyl-remote-mqtt/package.zip -d ${PREFIX}/var/db/baetyl/baetyl-remote-mqtt

	install -d -m 0755 ${PREFIX}/var/db/baetyl/baetyl-function-manager
	unzip -o baetyl-function-manager/package.zip -d ${PREFIX}/var/db/baetyl/baetyl-function-manager

	install -d -m 0755 ${PREFIX}/var/db/baetyl/baetyl-function-python27
	unzip -o baetyl-function-python/package27.zip -d ${PREFIX}/var/db/baetyl/baetyl-function-python27

	install -d -m 0755 ${PREFIX}/var/db/baetyl/baetyl-function-python36
	unzip -o baetyl-function-python/package36.zip -d ${PREFIX}/var/db/baetyl/baetyl-function-python36

	install -d -m 0755 ${PREFIX}/var/db/baetyl/baetyl-function-node85
	unzip -o baetyl-function-node/package85.zip -d ${PREFIX}/var/db/baetyl/baetyl-function-node85

	install -d -m 0755 ${PREFIX}/var/db/baetyl/baetyl-timer
	unzip -o baetyl-timer/package.zip -d ${PREFIX}/var/db/baetyl/baetyl-timer

	tar cf - -C example/native etc var | tar xvf - -C ${PREFIX}/

.PHONY: uninstall
uninstall:
	@rm -f ${PREFIX}/bin/baetyl
	@rm -rf ${PREFIX}/etc/baetyl
	@rm -rf ${PREFIX}/var/db/baetyl
	@rm -rf ${PREFIX}/var/log/baetyl
	@rm -rf ${PREFIX}/var/run/baetyl
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
	rm -f baetyl
	make -C baetyl-hub clean
	make -C baetyl-agent clean
	make -C baetyl-remote-mqtt clean
	make -C baetyl-function-manager clean
	make -C baetyl-function-python clean
	make -C baetyl-function-node clean
	make -C baetyl-timer clean

rebuild: clean all

generate:
	go generate ./...

image: clean
	for target in $(DEPLOY_TARGET) ; do \
		make -C baetyl-$$target image;\
	done

release: clean release-master release-image push-image release-manifest

release-master: clean
	# release linux 386
	env GOOS=linux GOARCH=386 make install PREFIX=__release_build/baetyl-$(VERSION)-linux-386
	cd __release_build/baetyl-$(VERSION)-linux-386 && zip -q -r ../../baetyl-$(VERSION)-linux-386.zip bin/
	make uninstall clean PREFIX=__release_build/baetyl-$(VERSION)-linux-386
	# release linux amd64
	env GOOS=linux GOARCH=amd64 make install PREFIX=__release_build/baetyl-$(VERSION)-linux-amd64
	cd __release_build/baetyl-$(VERSION)-linux-amd64 && zip -q -r ../../baetyl-$(VERSION)-linux-amd64.zip bin/
	make uninstall clean PREFIX=__release_build/baetyl-$(VERSION)-linux-amd64
	# release linux arm v7
	env GOOS=linux GOARCH=arm GOARM=7 make install PREFIX=__release_build/baetyl-$(VERSION)-linux-armv7
	cd __release_build/baetyl-$(VERSION)-linux-armv7 && zip -q -r ../../baetyl-$(VERSION)-linux-armv7.zip bin/
	make uninstall clean PREFIX=__release_build/baetyl-$(VERSION)-linux-armv7
	# release linux arm64
	env GOOS=linux GOARCH=arm64 make install PREFIX=__release_build/baetyl-$(VERSION)-linux-arm64
	cd __release_build/baetyl-$(VERSION)-linux-arm64 && zip -q -r ../../baetyl-$(VERSION)-linux-arm64.zip bin/
	make uninstall clean PREFIX=__release_build/baetyl-$(VERSION)-linux-arm64
	# release darwin amd64
	env GOOS=darwin GOARCH=amd64 make install PREFIX=__release_build/baetyl-$(VERSION)-darwin-amd64
	cd __release_build/baetyl-$(VERSION)-darwin-amd64 && zip -q -r ../../baetyl-$(VERSION)-darwin-amd64.zip bin/
	make uninstall PREFIX=__release_build/baetyl-$(VERSION)-darwin-amd64
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
		sed "s?__IMAGE_PREFIX__/?$(IMAGE_PREFIX)?g; s?__TAG__?$(VERSION)?g; s?__VERSION__?$(VERSION)?g; s?__TARGET__?baetyl-$$target?g;" manifest.yml.template > tmp/manifest-$$target-$(VERSION).yml;\
		./bin/manifest-tool-linux-amd64 --insecure --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-$$target-$(VERSION).yml;\
	done
	rm -rf tmp

release-manifest-latest:
	rm -rf tmp
	mkdir tmp
	for target in $(DEPLOY_TARGET) ; do \
		sed "s?__IMAGE_PREFIX__/?$(IMAGE_PREFIX)?g; s?__TAG__?latest?g; s?__VERSION__?$(VERSION)?g; s?__TARGET__?baetyl-$$target?g;" manifest.yml.template > tmp/manifest-$$target-latest.yml;\
		./bin/manifest-tool-linux-amd64 --insecure --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-$$target-latest.yml;\
	done
	rm -rf tmp

release-package: clean
	# Release modules' package -- linux armv7
	env GOOS=linux GOARCH=arm GOARM=7 make package
	for target in $(DEPLOY_TARGET) ; do \
		mv baetyl-$$target/package.zip ./baetyl-$$target-linux-armv7-$(VERSION).zip;\
	done
	make clean
	# Release modules' package -- linux amd64
	env GOOS=linux GOARCH=amd64 make package
	for target in $(DEPLOY_TARGET); do \
		mv baetyl-$$target/package.zip ./baetyl-$$target-linux-amd64-$(VERSION).zip;\
	done
	make clean
	# Release modules' package -- linux arm64
	env GOOS=linux GOARCH=arm64 make package
	for target in $(DEPLOY_TARGET); do \
		mv baetyl-$$target/package.zip ./baetyl-$$target-linux-arm64-$(VERSION).zip;\
	done
	make clean
	# Release modules' package -- linux 386
	env GOOS=linux GOARCH=386 make package
	for target in $(DEPLOY_TARGET); do \
		mv baetyl-$$target/package.zip ./baetyl-$$target-linux-386-$(VERSION).zip;\
	done
	make clean
	# Release modules' package -- darwin amd64
	env GOOS=darwin GOARCH=amd64 make package
	for target in $(DEPLOY_TARGET); do \
		mv baetyl-$$target/package.zip ./baetyl-$$target-darwin-amd64-$(VERSION).zip;\
	done
	make clean

push-image:
	for target in $(DEPLOY_TARGET); do \
		docker push $(IMAGE_PREFIX)baetyl-$$target:$(VERSION)-linux-amd64;\
		docker push $(IMAGE_PREFIX)baetyl-$$target:$(VERSION)-linux-arm64;\
		docker push $(IMAGE_PREFIX)baetyl-$$target:$(VERSION)-linux-armv7;\
		docker push $(IMAGE_PREFIX)baetyl-$$target:$(VERSION)-linux-386;\
	done
