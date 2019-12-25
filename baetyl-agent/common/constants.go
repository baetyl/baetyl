package common

type Resource string

const (
	Deployment  Resource = "deployment"
	Application Resource = "application"
	ConfigMap   Resource = "configMap"

	ShadowName      string = "BAETYL_SHADOW_NAME"
	ShadowNamespace string = "BAETYL_SHADOW_NAMESPACE"

	// ResourceType resource type
	ResourceType = "resourceType"
	// ResourceName resource name
	ResourceName = "resourceName"
)
