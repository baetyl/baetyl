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
	make -C openedge-timer clean

rebuild: clean all

generate:
	go generate ./...

image:
	make -C openedge-hub image
	make -C openedge-agent image
	make -C openedge-remote-mqtt image
	make -C openedge-function-manager image
	make -C openedge-function-python image
	make -C openedge-timer image

release: release-image
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
	# release linux arm
	env GOOS=linux GOARCH=arm make install PREFIX=__release_build/openedge-linux-arm-$(VERSION)
	tar czf openedge-linux-arm-$(VERSION).tar.gz -C __release_build/openedge-linux-arm-$(VERSION) bin etc var
	tar cjf openedge-linux-arm-$(VERSION).tar.bz2 -C __release_build/openedge-linux-arm-$(VERSION) bin etc var
	cd __release_build/openedge-linux-arm-$(VERSION) && zip -q -r ../../openedge-linux-arm-$(VERSION).zip bin/
	make uninstall clean PREFIX=__release_build/openedge-linux-arm-$(VERSION)
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
	env GOOS=linux GOARCH=arm make image IMAGE_SUFFIX="-linux-arm"
	make clean
	# linux-arm64 images release
	env GOOS=linux GOARCH=arm64 make image IMAGE_SUFFIX="-linux-arm64"
	make clean

# Need push built images first
release-manifest:
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
	sed "s/__REGISTRY__/$(REGISTRY)/g; s/__NAMESPACE__/$(NAMESPACE)/g; s/__VERSION__/$(VERSION)/g;" openedge-function-python/manifest27.yml.template > tmp/manifest-function-python27.yml
	./bin/manifest-tool-linux-amd64 --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-function-python27.yml
	# Push openedge-function-python36 manifest
	sed "s/__REGISTRY__/$(REGISTRY)/g; s/__NAMESPACE__/$(NAMESPACE)/g; s/__VERSION__/$(VERSION)/g;" openedge-function-python/manifest36.yml.template > tmp/manifest-function-python36.yml
	./bin/manifest-tool-linux-amd64 --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-function-python36.yml
	# Push openedge-remote-mqtt manifest
	sed "s/__REGISTRY__/$(REGISTRY)/g; s/__NAMESPACE__/$(NAMESPACE)/g; s/__VERSION__/$(VERSION)/g;" openedge-remote-mqtt/manifest.yml.template > tmp/manifest-remote-mqtt.yml
	./bin/manifest-tool-linux-amd64 --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-remote-mqtt.yml
	# Push openedge-timer manifest
	sed "s/__REGISTRY__/$(REGISTRY)/g; s/__NAMESPACE__/$(NAMESPACE)/g; s/__VERSION__/$(VERSION)/g;" openedge-timer/manifest.yml.template > tmp/manifest-timer.yml
	./bin/manifest-tool-linux-amd64 --username=$(USERNAME) --password=$(PASSWORD) push from-spec tmp/manifest-timer.yml

	rm -rf tmp

release-package:
	# Release modules' package -- linux arm
	env GOOS=linux GOARCH=arm make package
	mv openedge-agent/package.zip ./openedge-agent-linux-arm-$(VERSION).zip
	mv openedge-hub/package.zip ./openedge-hub-linux-arm-$(VERSION).zip
	mv openedge-remote-mqtt/package.zip ./openedge-remote-mqtt-linux-arm-$(VERSION).zip
	mv openedge-function-manager/package.zip ./openedge-function-manager-linux-arm-$(VERSION).zip
	mv openedge-function-python/package27.zip ./openedge-function-python27-linux-arm-$(VERSION).zip
	mv openedge-function-python/package36.zip ./openedge-function-python36-linux-arm-$(VERSION).zip
	mv openedge-timer/package.zip ./openedge-timer-linux-arm-$(VERSION).zip
	# Release modules' package -- linux amd64
	env GOOS=linux GOARCH=amd64 make package
	mv openedge-agent/package.zip ./openedge-agent-linux-amd64-$(VERSION).zip
	mv openedge-hub/package.zip ./openedge-hub-linux-amd64-$(VERSION).zip
	mv openedge-remote-mqtt/package.zip ./openedge-remote-mqtt-linux-amd64-$(VERSION).zip
	mv openedge-function-manager/package.zip ./openedge-function-manager-linux-amd64-$(VERSION).zip
	mv openedge-function-python/package27.zip ./openedge-function-python27-linux-amd64-$(VERSION).zip
	mv openedge-function-python/package36.zip ./openedge-function-python36-linux-amd64-$(VERSION).zip
	mv openedge-timer/package.zip ./openedge-timer-linux-amd64-$(VERSION).zip
	# Release modules' package -- linux arm64
	env GOOS=linux GOARCH=arm64 make package
	mv openedge-agent/package.zip ./openedge-agent-linux-arm64-$(VERSION).zip
	mv openedge-hub/package.zip ./openedge-hub-linux-arm64-$(VERSION).zip
	mv openedge-remote-mqtt/package.zip ./openedge-remote-mqtt-linux-arm64-$(VERSION).zip
	mv openedge-function-manager/package.zip ./openedge-function-manager-linux-arm64-$(VERSION).zip
	mv openedge-function-python/package27.zip ./openedge-function-python27-linux-arm64-$(VERSION).zip
	mv openedge-function-python/package36.zip ./openedge-function-python36-linux-arm64-$(VERSION).zip
	mv openedge-timer/package.zip ./openedge-timer-linux-arm64-$(VERSION).zip
	# Release modules' package -- linux 386
	env GOOS=linux GOARCH=386 make package
	mv openedge-agent/package.zip ./openedge-agent-linux-386-$(VERSION).zip
	mv openedge-hub/package.zip ./openedge-hub-linux-386-$(VERSION).zip
	mv openedge-remote-mqtt/package.zip ./openedge-remote-mqtt-linux-386-$(VERSION).zip
	mv openedge-function-manager/package.zip ./openedge-function-manager-linux-386-$(VERSION).zip
	mv openedge-function-python/package27.zip ./openedge-function-python27-linux-386-$(VERSION).zip
	mv openedge-function-python/package36.zip ./openedge-function-python36-linux-386-$(VERSION).zip
	mv openedge-timer/package.zip ./openedge-timer-linux-386-$(VERSION).zip
	# Release modules' package -- darwin amd64
	env GOOS=darwin GOARCH=amd64 make package
	mv openedge-agent/package.zip ./openedge-agent-darwin-amd64-$(VERSION).zip
	mv openedge-hub/package.zip ./openedge-hub-darwin-amd64-$(VERSION).zip
	mv openedge-remote-mqtt/package.zip ./openedge-remote-mqtt-darwin-amd64-$(VERSION).zip
	mv openedge-function-manager/package.zip ./openedge-function-manager-darwin-amd64-$(VERSION).zip
	mv openedge-function-python/package27.zip ./openedge-function-python27-darwin-amd64-$(VERSION).zip
	mv openedge-function-python/package36.zip ./openedge-function-python36-darwin-amd64-$(VERSION).zip
	mv openedge-timer/package.zip ./openedge-timer-darwin-amd64-$(VERSION).zip
