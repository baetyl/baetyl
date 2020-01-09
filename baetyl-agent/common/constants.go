package common

type Resource string

const (
	// KeyContextNamespace the key of namespace in context
	KeyContextNamespace = "namespace"

	Deployment  Resource = "deployment"
	Application Resource = "application"
	Config      Resource = "config"
	Batch       Resource = "batch"

	NodeName      string = "BAETYL_NODE_NAME"
	NodeNamespace string = "BAETYL_NODE_NAMESPACE"
	NodeID        string = "BAETYL_NODE_ID"

	// HeaderKeyNodeNamespace header key of node namespace
	HeaderKeyNodeNamespace = "node-namespace"
	// HeaderKeyNodeName header key of node name
	HeaderKeyNodeName = "node-name"
)
