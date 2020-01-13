package common

type Resource string
type Proof string

const (
	// KeyContextNamespace the key of namespace in context
	KeyContextNamespace = "namespace"

	Deployment  Resource = "deployment"
	Application Resource = "application"
	Config      Resource = "config"
	Batch       Resource = "batch"
	Node        Resource = "node"

	HostID Proof = "hostID"
	CPU    Proof = "cpu"
	MAC    Proof = "mac"
	SN     Proof = "sn"

	SNPath = "var/lib/baetyl/sn/"

	NodeName      string = "BAETYL_NODE_NAME"
	NodeNamespace string = "BAETYL_NODE_NAMESPACE"
	NodeID        string = "BAETYL_NODE_ID"

	BatchName      string = "BAETYL_BATCH_NAME"
	BatchNamespace string = "BAETYL_BATCH_NAMESPACE"

	// HeaderKeyNodeNamespace header key of node namespace
	HeaderKeyNodeNamespace = "node-namespace"
	// HeaderKeyNodeName header key of node name
	HeaderKeyNodeName = "node-name"
	// HeaderKeyBatchNamespace header key of batch namespace
	HeaderKeyBatchNamespace = "batch-namespace"
	// HeaderKeyBatchName header key of batch name
	HeaderKeyBatchName = "batch-name"
)
