module github.com/baetyl/baetyl-core

go 1.13

replace (
	github.com/docker/docker => github.com/docker/engine v0.0.0-20191007211215-3e077fc8667a
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc6.0.20190307181833-2b18fe1d885e
)

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/baetyl/baetyl-go v0.1.24
	github.com/containerd/containerd v1.3.0 // indirect
	github.com/docker/docker v0.7.3-0.20190327010347-be7ac8be2ae0
	github.com/docker/go-units v0.4.0
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/jinzhu/copier v0.0.0-20190924061706-b57f9002281a
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/timshannon/bolthold v0.0.0-20200310154430-7be3f3bd401d
	go.etcd.io/bbolt v1.3.3
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kubectl v0.17.4
	k8s.io/metrics v0.17.4
)
