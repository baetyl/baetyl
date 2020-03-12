package common

type Resource string
type Proof string

const (
	DefaultConfFile = "conf/config.yml"

	KeyContextNamespace = "namespace"

	Deployment    Resource = "deployment"
	Application   Resource = "application"
	Configuration Resource = "configuration"
	Node          Resource = "node"

	HostID Proof = "hostID"
	CPU    Proof = "cpu"

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
	DefaultNamespace   = "baetyl-edge"
	ZipCompression     = "zip"

	InternalEventTopic = "mqtt/event"

	DefaultAppsKey = "apps"

	APPVersionMapping = "app/mapping"
)
