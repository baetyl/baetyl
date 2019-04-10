package engine

import (
	openedge "github.com/baidu/openedge/sdk/openedge-go"
)

// Service interfaces of service
type Service interface {
	Name() string
	Stats() openedge.ServiceStatus
	Start() error
	Stop()
	ReportInstance(instanceName string, stats map[string]interface{}) error
	StartInstance(instanceName string, dynamicConfig map[string]string) error
	StopInstance(instanceName string) error
}
