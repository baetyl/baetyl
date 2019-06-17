package engine

import openedge "github.com/baidu/openedge/sdk/openedge-go"

// Service interfaces of service
type Service interface {
	Name() string
	Engine() Engine
	RestartPolicy() openedge.RestartPolicyInfo
	Start() error
	Stop()
	Stats()
	StartInstance(instanceName string, dynamicConfig map[string]string) error
	StopInstance(instanceName string) error
}
