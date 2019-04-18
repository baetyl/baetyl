package engine

// ServicesStats stats of all services
type ServicesStats map[string]InstancesStats

// InstancesStats stats of all instances of the service
type InstancesStats map[string]PartialStats

// PartialStats partial stats of the instance
type PartialStats map[string]interface{}

// NewPartialStatsByStatus creates a new stats by status
func NewPartialStatsByStatus(status string) PartialStats {
	return PartialStats{
		KeyStatus: status,
	}
}

// InfoStats interfaces of the storage of info and stats
type InfoStats interface {
	AddInstanceStats(serviceName, instanceName string, partialStats PartialStats)
	DelInstanceStats(serviceName, instanceName string)
}
