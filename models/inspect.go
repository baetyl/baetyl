package models

import "time"

// Report report
type Report map[string]interface{}

// Desire desire
type Desire map[string]interface{}

// ServiceInfo service info
type ServiceInfo struct {
	Name       string                   `yaml:"name,omitempty" json:"name,omitempty"`
	Container  Container                `yaml:"container,omitempty" json:"container,omitempty"`
	Usage      map[string]*ResourceInfo `yaml:"usage,omitempty" json:"usage,omitempty"`
	Status     string                   `yaml:"status,omitempty" json:"status,omitempty"`
	Cause      string                   `yaml:"cause,omitempty" json:"cause,omitempty"`
	CreateTime time.Time                `yaml:"createTime,omitempty" json:"createTime,omitempty"`
}

// Container container info
type Container struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	ID   string `yaml:"id,omitempty" json:"id,omitempty"`
}

// VolumeInfo volume info
type VolumeInfo struct {
	Name    string `yaml:"name,omitempty" json:"name,omitempty"`
	Version string `yaml:"version,omitempty" json:"version,omitempty"`
}

// ResourceInfo resource info
type ResourceInfo struct {
	Name        string `yaml:"name,omitempty" json:"name,omitempty"`
	Value       string `yaml:"value,omitempty" json:"value,omitempty"`
	UsedPercent string `yaml:"usedPercent,omitempty" json:"usedPercent,omitempty"`
}

type NodeStats struct {
	Usage    map[string]*ResourceInfo `yaml:"usage,omitempty" json:"usage,omitempty"`
	Capacity map[string]*ResourceInfo `yaml:"capacity,omitempty" json:"capacity,omitempty"`
}

// NodeInfo node info
type NodeInfo struct {
	Hostname         string `yaml:"hostname,omitempty" json:"hostname,omitempty"`
	Address          string `yaml:"address,omitempty" json:"address,omitempty"`
	Arch             string `yaml:"arch,omitempty" json:"arch,omitempty"`
	KernelVersion    string `yaml:"kernelVer,omitempty" json:"kernelVer,omitempty"`
	OS               string `yaml:"os,omitempty" json:"os,omitempty"`
	ContainerRuntime string `yaml:"containerRuntime,omitempty" json:"containerRuntime"`
	MachineID        string `yaml:"machineID,omitempty" json:"machineID"`
	SystemUUID       string `yaml:"systemUUID,omitempty" json:"systemUUID"`
	OSImage          string `yaml:"osImage,omitempty" json:"osImage"`
}

// AppStats app stats
type AppStats struct {
	Name         string                  `yaml:"name" json:"name"`
	Version      string                  `yaml:"version,omitempty" json:"version,omitempty"`
	Status       string                  `yaml:"status,omitempty" json:"status,omitempty"`
	Cause        string                  `yaml:"cause,omitempty" json:"cause,omitempty"`
	ServiceInfos map[string]*ServiceInfo `yaml:"services,omitempty" json:"services,omitempty"`
	VolumeInfos  map[string]*VolumeInfo  `yaml:"volumes,omitempty" json:"volumes,omitempty"`
}
