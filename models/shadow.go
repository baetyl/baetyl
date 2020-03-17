package models

import "time"

// Shadow the mode of shadow
type Shadow struct {
	Name              string                 `json:"name,omitempty"`
	Namespace         string                 `json:"namespace,omitempty"`
	Version           string                 `json:"version,omitempty"`
	CreationTimestamp time.Time              `json:"creationTimestamp,omitempty"`
	Labels            map[string]string      `json:"labels,omitempty"`
	Reported          map[string]interface{} `json:"reported,omitempty"`
	Desired           map[string]interface{} `json:"desired,omitempty"`
}
