package models

type Shadow struct {
	Name       string            `json:"name,omitempty"`
	Namespace  string            `json:"namespace,omitempty"`
	Version    string            `json:"version,omitempty"`
	Reported   interface{}       `json:"reported,omitempty"`
	Desired    interface{}       `json:"desired,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}
