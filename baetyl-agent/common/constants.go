package common

type Resource string

const (
	Deployment  Resource = "deployment"
	Application Resource = "application"
	Config      Resource = "config"
	Batch       Resource = "batch"

	NodeName      string = "BAETYL_NODE_NAME"
	NodeNamespace string = "BAETYL_NODE_NAMESPACE"
	NodeID        string = "BAETYL_NODE_ID"

	// ResourceType resource type
	ResourceType = "resourceType"
	// ResourceName resource name
	ResourceName = "resourceName"
)
