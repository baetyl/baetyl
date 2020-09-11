package ami

// the host paths are fixed currently
// TODO: to find the hostpath from app volumes, does not hardcode
// Maybe you can use the environment variables to determine the host paths.
// All paths are set during installation, and host path mapping is set by macro in application model
// to avoid environment difference in application model
// baetyl-488 [Task] 应用模型支持边缘宏变量，避免在应用模型上出现环境差异的配置
const (
	RunHostPath    = "/var/lib/baetyl/run"
	LogHostPath    = "/var/lib/baetyl/log"
	HostHostPath   = "/var/lib/baetyl/host"
	ObjectHostPath = "/var/lib/baetyl/object"
)
