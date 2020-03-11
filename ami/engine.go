package ami

import (
	"github.com/baetyl/baetyl-core/models"
	"github.com/jinzhu/copier"
	v1 "k8s.io/api/core/v1"
)

func configurationToConfigMap(config *models.Configuration) (*v1.ConfigMap, error) {
	configMap := &v1.ConfigMap{}
	err := copier.Copy(configMap, config)
	if err != nil {
		return nil, err
	}
	return configMap, nil
}



