package master

import "github.com/docker/docker/api/types"

// ModuleStats stats of module
type ModuleStats struct {
	types.Stats `json:",inline,omitempt"`
	Name        string `json:"name,omitempty"`
	Type        string `json:"type,omitempty"`
	Error       string `json:"error,omitempty"`
	Status      string `json:"status,omitempty"`
	StartedAt   string `json:"start_time,omitempty"`
	FinishedAt  string `json:"finish_time,omitempty"`
}

// Stats all stats
type Stats struct {
	Info map[string]interface{} `json:"info,omitempty"`
}

// NewStats creates a stats
func NewStats() *Stats {
	return &Stats{
		Info: make(map[string]interface{}),
	}
}
