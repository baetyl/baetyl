package openedge

import (
	"encoding/json"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/protocol/mqtt"
)

// Config of service
type Config struct {
	Name   string          `yaml:"name" json:"name"`
	Hub    mqtt.ClientInfo `yaml:"hub" json:"hub"`
	Logger logger.LogInfo  `yaml:"logger" json:"logger"`
}

// DatasetInfo dataset for master
type DatasetInfo struct {
	Name    string `yaml:"name" json:"name" validate:"nonzero"`
	Version string `yaml:"verion" json:"version"`
	URL     string `yaml:"url" json:"url"`
	MD5     string `yaml:"mds" json:"mds"`
}

//NewDatasetInfoFromBytes creates dataset info by bytes
func NewDatasetInfoFromBytes(d []byte) (*DatasetInfo, error) {
	dataset := new(DatasetInfo)
	err := json.Unmarshal(d, dataset)
	if err != nil {
		return nil, err
	}
	return dataset, nil
}
