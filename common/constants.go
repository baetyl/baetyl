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

	SN        Proof = "sn"
	Input     Proof = "input"
	HostName  Proof = "hostName"
	MachineID Proof = "machineID"

	NodeName       string = "BAETYL_NODE_NAME"
	NodeNamespace  string = "BAETYL_NODE_NAMESPACE"
	BatchName      string = "BAETYL_BATCH_NAME"
	BatchNamespace string = "BAETYL_BATCH_NAMESPACE"
	NodeID         string = "BAETYL_NODE_ID"

	// HeaderKeyNodeNamespace header key of node namespace
	HeaderKeyNodeNamespace = "node-namespace"
	// HeaderKeyNodeName header key of node name
	HeaderKeyNodeName = "node-name"
	// HeaderKeyBatchNamespace header key of batch namespace
	HeaderKeyBatchNamespace = "batch-namespace"
	// HeaderKeyBatchName header key of batch name
	HeaderKeyBatchName = "batch-name"
	// StorageObjectPrefix prefix of storage object
	PrefixConfigObject = "_object_"
	DefaultNamespace   = "default"
	ZipCompression     = "zip"

	InternalEventTopic = "mqtt/event"

	DefaultAppsKey = "apps"

	APPVersionMapping = "app/mapping"
)
