package kube

import (
	"encoding/json"
	"github.com/baetyl/baetyl-core/common"
)

func (k *kubeEngine) UpdateApp(apps map[string]string) error {
	// TODO start a routine to get status from api server
	res, err := json.Marshal(apps)
	if err != nil {
		return err
	}
	err = k.driver.Update([]byte(common.DefaultAppsKey), res)
	if err != nil {
		return err
	}
	return nil
}
