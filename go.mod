module github.com/baetyl/baetyl-core

go 1.13

replace (
	github.com/256dpi/gomqtt => github.com/256dpi/gomqtt v0.12.2
	github.com/docker/docker => github.com/docker/engine v0.0.0-20191007211215-3e077fc8667a
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc6.0.20190307181833-2b18fe1d885e
)

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/StackExchange/wmi v0.0.0-20190523213609-cbe669659 // indirect
	github.com/baetyl/baetyl v0.0.0-20190912105404-a936550b7992
	github.com/baetyl/baetyl-go v0.1.8
	github.com/docker/go-units v0.4.0
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/shirou/gopsutil v2.20.2+incompatible // indirect
	github.com/sirupsen/logrus v1.4.2 // indirect
	k8s.io/api v0.0.0-20190620084959-7cf5895f2711
	k8s.io/apimachinery v0.0.0-20190612205821-1799e75a0719
	k8s.io/client-go v12.0.0+incompatible
)
