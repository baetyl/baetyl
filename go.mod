module github.com/baetyl/baetyl/v2

go 1.13

replace github.com/kardianos/service => github.com/baetyl/service v0.0.0-20200910124134-20fdd363fbd5

require (
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/baetyl/baetyl-go/v2 v2.1.10
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/golang/mock v1.3.1
	github.com/jinzhu/copier v0.1.0
	github.com/kardianos/service v1.1.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/pkg/errors v0.9.1
	github.com/qiangxue/fasthttp-routing v0.0.0-20160225050629-6ccdc2a18d87
	github.com/shirou/gopsutil v2.20.5+incompatible
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.5.1
	github.com/timshannon/bolthold v0.0.0-20200310154430-7be3f3bd401d
	github.com/valyala/fasthttp v1.9.0
	go.etcd.io/bbolt v1.3.3
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.17.8
	k8s.io/apimachinery v0.17.8
	k8s.io/client-go v0.17.8
	k8s.io/kubectl v0.17.8
	k8s.io/metrics v0.17.8
)
