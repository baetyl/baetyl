module github.com/baetyl/baetyl/v2

go 1.13

replace github.com/kardianos/service => github.com/baetyl/service v0.0.0-20220214084200-533e95169992

require (
	github.com/baetyl/baetyl-go/v2 v2.2.4-0.20220124101541-b5b5f0c98390
	github.com/golang/mock v1.3.1
	github.com/imdario/mergo v0.3.5
	github.com/jinzhu/copier v0.1.0
	github.com/kardianos/service v1.2.1
	github.com/mitchellh/mapstructure v1.1.2
	github.com/pkg/errors v0.9.1
	github.com/qiangxue/fasthttp-routing v0.0.0-20160225050629-6ccdc2a18d87
	github.com/shirou/gopsutil v3.21.9+incompatible
	github.com/shirou/gopsutil/v3 v3.21.9
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.7.0
	github.com/timshannon/bolthold v0.0.0-20200310154430-7be3f3bd401d
	github.com/valyala/fasthttp v1.9.0
	go.etcd.io/bbolt v1.3.6
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.17.8
	k8s.io/apimachinery v0.17.8
	k8s.io/client-go v0.17.8
	k8s.io/kubectl v0.17.8
	k8s.io/metrics v0.17.8
)
