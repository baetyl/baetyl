package openedge

import "time"

// Inspect openedge information inspected
type Inspect struct {
	Error    string    `json:"error,omitempt"`
	Time     time.Time `json:"time,omitempt"`
	Platform Platform  `json:"platform,omitempt"`
	HostInfo HostInfo  `json:"host_info,omitempt"`
	Services Services  `json:"services,omitempt"`
	// Datasets []DatasetStatus
	// Volumes  []VolumeStatus
}

// Platform platform information
type Platform struct {
	Mode        string `json:"mode,omitempt"`
	GoVersion   string `json:"go_version,omitempt"`
	BinVersion  string `json:"bin_version,omitempt"`
	ConfVersion string `json:"conf_version,omitempt"`
}

// HostInfo host information
type HostInfo map[string]interface{}

// Services all services' information
type Services []ServiceStatus

// ServiceStatus service status
type ServiceStatus struct {
	Name      string           `json:"name,omitempt"`
	Instances []InstanceStatus `json:"instances,omitempt"`
}

// InstanceStatus service instance status
type InstanceStatus map[string]interface{}

// NewInspect create a new information inspected
func NewInspect() *Inspect {
	return &Inspect{
		HostInfo: HostInfo{},
		Services: Services{},
	}
}

// NewServiceStatus create a new service status
func NewServiceStatus(name string) ServiceStatus {
	return ServiceStatus{
		Name:      name,
		Instances: []InstanceStatus{},
	}
}
