package openedge

// Inspect openedge information inspected
type Inspect struct {
	Error     string   `json:"error,omitempt"`
	Timestamp int64    `json:"timestamp,omitempt"`
	Platform  Platform `json:"platform,omitempt"`
	HostInfo  HostInfo `json:"host_info,omitempt"`
	Services  Services `json:"services,omitempt"`
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
type ServiceStatus []InstanceStatus

// InstanceStatus service instance status
type InstanceStatus map[string]interface{}

// NewInspect create a new information inspected
func NewInspect() *Inspect {
	return &Inspect{
		HostInfo: HostInfo{},
		Services: Services{},
	}
}
