module github.com/baetyl/baetyl

go 1.13

replace (
	github.com/256dpi/gomqtt => github.com/256dpi/gomqtt v0.12.2
	github.com/docker/docker => github.com/docker/engine v0.0.0-20191007211215-3e077fc8667a
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc6.0.20190307181833-2b18fe1d885e
)

require (
	github.com/256dpi/gomqtt v0.13.0
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/baetyl/baetyl-go v0.0.0-20191218020428-5414115311d0
	github.com/containerd/containerd v1.3.0
	github.com/creasty/defaults v1.3.0
	github.com/deckarep/golang-set v1.7.1
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v0.0.0-00010101000000-000000000000
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-units v0.4.0
	github.com/etcd-io/bbolt v1.3.3
	github.com/fortytw2/leaktest v1.3.0 // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/goinggo/mapstructure v0.0.0-20140717182941-194205d9b4a9
	github.com/golang/protobuf v1.3.2
	github.com/gorilla/mux v1.7.3
	github.com/inconshreveable/go-update v0.0.0-20160112193335-8152e7eb6ccf
	github.com/jolestar/go-commons-pool v2.0.0+incompatible
	github.com/jpillora/backoff v1.0.0
	github.com/mattn/go-sqlite3 v1.11.0
	github.com/mholt/archiver v3.1.1+incompatible
	github.com/mitchellh/mapstructure v1.1.2
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/opencontainers/runc v0.0.0-00010101000000-000000000000 // indirect
	github.com/opencontainers/runtime-spec v1.0.1 // indirect
	github.com/orcaman/concurrent-map v0.0.0-20190826125027-8c72a8bb44f6
	github.com/pierrec/lz4 v2.4.0+incompatible // indirect
	github.com/shirou/gopsutil v2.19.11+incompatible
	github.com/shirou/w32 v0.0.0-20160930032740-bb4de0191aa4 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.4.0
	go.etcd.io/bbolt v1.3.3 // indirect
	gocv.io/x/gocv v0.21.0
	golang.org/x/net v0.0.0-20191119073136-fc4aabc6c914
	golang.org/x/sys v0.0.0-20191223224216-5a3cf8467b4e // indirect
	golang.org/x/text v0.3.2 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	google.golang.org/grpc v1.25.1
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637
	gopkg.in/validator.v2 v2.0.0-20191107172027-c3144fdedc21
	gopkg.in/yaml.v2 v2.2.4
	gotest.tools v2.2.0+incompatible // indirect
)
