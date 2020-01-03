package common

type Resource string

const (
	Deployment  Resource = "deployment"
	Application Resource = "application"
	Config      Resource = "config"

	NodeName      string = "BAETYL_NODE_NAME"
	NodeNamespace string = "BAETYL_NODE_NAMESPACE"
	NodeID        string = "BAETYL_NODE_ID"

	// ResourceType resource type
	ResourceType = "resourceType"
	// ResourceName resource name
	ResourceName = "resourceName"
	// ResourceVersion resource version
	ResourceVersion = "resourceVersion"
)
