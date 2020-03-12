package kube

import (
	"github.com/baetyl/baetyl-core/common"
)

func (k *kubeEngine) UpdateApp(apps map[string]string) error {
	// TODO start a routine to get status from api server
	return k.store.Upsert(common.DefaultAppsKey, apps)
}
