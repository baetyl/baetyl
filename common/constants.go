package common

type Resource string
type Proof string

const (
	KeyContextNamespace = "namespace"

	Deployment    Resource = "deployment"
	Application   Resource = "application"
	Configuration Resource = "configuration"
	Node          Resource = "node"
	Secret        Resource = "secret"

	SN         Proof = "sn"
	Input      Proof = "input"
	HostName   Proof = "hostName"
	MachineID  Proof = "machineID"
	SystemUUID Proof = "systemUUID"

	DefaultSNPath = "var/lib/baetyl/sn"
	KeyHostName   = "NODE_NAME"

	KeyActivateDataBatchName        = "batchName"
	KeyActivateDataBatchNamespace   = "batchNamespace"
	KeyActivateDataSecurityType     = "securityType"
	KeyActivateDataSecurityKey      = "securityKey"
	KeyActivateDataFingerprintValue = "fingerprintValue"
	KeyActivateResNodeName          = "nodeName"
	KeyActivateResNodeNamespace     = "namespace"
	KeyActivateResCA                = "ca"
	KeyActivateResCert              = "cert"
	KeyActivateResKey               = "key"
	KeyActivateResName              = "name"

	NodeName       = "BAETYL_NODE_NAME"
	NodeNamespace  = "BAETYL_NODE_NAMESPACE"
	BatchName      = "BAETYL_BATCH_NAME"
	BatchNamespace = "BAETYL_BATCH_NAMESPACE"
	NodeID         = "BAETYL_NODE_ID"

	// HeaderKeyNodeNamespace header key of node namespace
	HeaderKeyNodeNamespace = "node-namespace"
	// HeaderKeyNodeName header key of node name
	HeaderKeyNodeName = "node-name"
	// StorageObjectPrefix prefix of storage object
	PrefixConfigObject = "_object_"
	DefaultNamespace   = "default"
	ZipCompression     = "zip"

	DefaultAppsKey = "apps"

	SyncDesireEvent = "sync/desire"
	SyncReportEvent = "sync/report"
	EngineAppEvent  = "engine/app"

	EventCenterLimit = 20
)
