package engine

import baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"

// Service interfaces of service
type Service interface {
	Name() string
	Engine() Engine
	RestartPolicy() baetyl.RestartPolicyInfo
	Start() error
	Stop()
	Stats()
	StartInstance(instanceName string, dynamicConfig map[string]string) error
	StopInstance(instanceName string) error
}
