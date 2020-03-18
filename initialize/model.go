package initialize

import "github.com/baetyl/baetyl-go/utils"

type BackwardInfo struct {
	NodeName  string            `yaml:"nodeName,omitempty" json:"nodeName,omitempty"`
	Namespace string            `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	Cert      utils.Certificate `yaml:"cert,omitempty" json:"cert,omitempty"`
}

type ForwardInfo struct {
	BatchName        string            `yaml:"batchName,omitempty" json:"batchName,omitempty"`
	Namespace        string            `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	FingerprintValue string            `yaml:"fingerprintValue,omitempty" json:"fingerprintValue,omitempty"`
	SecurityType     string            `yaml:"securityType,omitempty" json:"securityType,omitempty"`
	SecurityValue    string            `yaml:"securityValue,omitempty" json:"securityValue,omitempty"`
	PenetrateData    map[string]string `yaml:"penetrateData,omitempty" json:"penetrateData,omitempty"`
}
