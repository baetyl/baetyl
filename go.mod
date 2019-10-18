module github.com/baetyl/baetyl

go 1.13

replace (
	github.com/docker/docker => github.com/docker/engine v0.0.0-20191007211215-3e077fc8667a
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.1-0.20190307181833-2b18fe1d885e
)

require (
	github.com/256dpi/gomqtt v0.12.2
	github.com/containerd/containerd v1.3.0 // indirect
	github.com/creasty/defaults v1.3.0
	github.com/deckarep/golang-set v1.7.1
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v0.0.0-00010101000000-000000000000
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-units v0.4.0
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/etcd-io/bbolt v1.3.3
	github.com/goburrow/modbus v0.1.0
	github.com/goburrow/serial v0.1.0
	github.com/gogo/protobuf v1.3.0 // indirect
	github.com/golang/protobuf v1.3.2
	github.com/golang/snappy v0.0.1 // indirect
	github.com/gorilla/mux v1.7.3
	github.com/inconshreveable/go-update v0.0.0-20160112193335-8152e7eb6ccf
	github.com/jolestar/go-commons-pool v2.0.0+incompatible
	github.com/jpillora/backoff v1.0.0
	github.com/mholt/archiver v3.1.1+incompatible
	github.com/nwaples/rardecode v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/opencontainers/runc v0.0.0-00010101000000-000000000000 // indirect
	github.com/opencontainers/runtime-spec v1.0.1 // indirect
	github.com/orcaman/concurrent-map v0.0.0-20190826125027-8c72a8bb44f6
	github.com/pierrec/lz4 v2.3.0+incompatible // indirect
	github.com/pkg/errors v0.8.1 // indirect
	github.com/shirou/gopsutil v2.19.9+incompatible
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.2.2
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	gocv.io/x/gocv v0.21.0
	golang.org/x/net v0.0.0-20191011234655-491137f69257
	google.golang.org/grpc v1.24.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637
	gopkg.in/validator.v2 v2.0.0-20191008145730-5614e8810ea7
	gopkg.in/yaml.v2 v2.2.4
)
