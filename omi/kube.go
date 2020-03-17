package omi

import (
	"github.com/baetyl/baetyl-core/common"
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/models"
	"github.com/baetyl/baetyl-core/utils"
	"github.com/jinzhu/copier"
	bh "github.com/timshannon/bolthold"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/reference"
	"k8s.io/kubectl/pkg/scheme"
	"time"
)

type kubeModel struct {
	cli      *Client
	store    *bh.Store
	nodeName string
}

func NewKubeModel(cfg config.APIServer, store *bh.Store) (Model, error) {
	cli, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &kubeModel{cli: cli, store: store}, nil
}

func (k *kubeModel) CollectInfo(apps map[string]string) (map[string]interface{}, error) {
	nodeInfo, err := k.collectNodeInfo()
	if err != nil {
		return nil, err
	}
	appStatus, err := k.collectAppStatus(apps)
	if err != nil {
		return nil, err
	}

	info := map[string]interface{}{
		"time": time.Now(),
		"node": nodeInfo,
		"apps": appStatus,
	}
	return info, nil
}

func (k *kubeModel) collectNodeInfo() (*models.NodeInfo, error) {
	node, err := k.cli.Core.Nodes().Get(k.nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	ni := node.Status.NodeInfo
	nodeInfo := &models.NodeInfo{
		Arch:             ni.Architecture,
		KernelVersion:    ni.KernelVersion,
		OS:               ni.OperatingSystem,
		ContainerRuntime: ni.ContainerRuntimeVersion,
		MachineID:        ni.MachineID,
		OSImage:          ni.OSImage,
	}

	for _, addr := range node.Status.Addresses {
		if addr.Type == v1.NodeInternalIP {
			nodeInfo.Address = addr.Address
		} else if addr.Type == v1.NodeHostName {
			nodeInfo.Hostname = addr.Address
		}
	}
	for res, quantity := range node.Status.Capacity {
		nodeInfo.Capacity[string(res)] = models.ResourceInfo{
			Name:  string(res),
			Value: quantity.String(),
		}
	}
	return nodeInfo, nil
}

func (k *kubeModel) collectAppStatus(apps map[string]string) (map[string]*models.AppStatus, error) {
	res := map[string]*models.AppStatus{}
	for name, ver := range apps {
		status := &models.AppStatus{
			Name:    name,
			Version: ver,
		}
		deploy, err := k.cli.App.Deployments(k.cli.Namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		ref, err := reference.GetReference(scheme.Scheme, deploy)
		events, _ := k.cli.Core.Events(k.cli.Namespace).Search(scheme.Scheme, ref)
		for _, e := range events.Items {
			if e.Type == "Warning" {
				status.Cause += e.Message + "\n"
			}
		}
		status.ServiceInfos = map[string]models.ServiceInfo{}
		status.VolumeInfos = map[string]models.VolumeInfo{}
		res[name] = status
	}
	return res, nil
}

func (k *kubeModel) collectService(name string) (*models.ServiceInfo, error) {
	return nil, nil
}

func (k *kubeModel) collectVolumeInfo(name string) (*models.VolumeInfo, error) {
	return nil, nil
}

func (k *kubeModel) ApplyApplications(info map[string]string) error {
	var apps config.AppsVersionResource
	err := k.store.Get(common.DefaultAppsKey, &apps)
	if err != nil {
		return err
	}
	// TODO: why to get here, then to update in update.go?
	return k.UpdateApp(apps.Value)
}

func (k *kubeModel) UpdateApp(apps map[string]string) error {
	deploys := map[string]*appv1.Deployment{}
	var services []*v1.Service
	configs := map[string]*v1.ConfigMap{}
	for name, ver := range apps {
		var app models.Application
		key := utils.MakeKey(common.Application, name, ver)
		err := k.store.Get(key, &app)
		if err != nil {
			return err
		}
		deploy, svcs, err := toDeployAndService(&app)
		if err != nil {
			return err
		}
		deploys[deploy.Name] = deploy
		services = append(services, svcs...)
		for _, v := range app.Volumes {
			if cfg := v.Configuration; cfg != nil {
				key := utils.MakeKey(common.Configuration, cfg.Name, cfg.Version)
				var config models.Configuration
				err := k.store.Get(key, &config)
				if err != nil {
					return err
				}
				configMap, err := toConfigMap(&config)
				configs[config.Name] = configMap
			}
		}
	}

	configMapInterface := k.cli.Core.ConfigMaps(k.cli.Namespace)
	for _, cfg := range configs {
		_, err := configMapInterface.Get(cfg.Name, metav1.GetOptions{})
		if err == nil {
			_, err := configMapInterface.Update(cfg)
			if err != nil {
				return err
			}
		} else {
			_, err := configMapInterface.Create(cfg)
			if err != nil {
				return err
			}
		}
	}

	deployInterface := k.cli.App.Deployments(k.cli.Namespace)
	for _, d := range deploys {
		_, err := deployInterface.Get(d.Name, metav1.GetOptions{})
		if err == nil {
			_, err := deployInterface.Update(d)
			if err != nil {
				return err
			}
		} else {
			_, err := deployInterface.Create(d)
			if err != nil {
				return err
			}
		}
	}

	serviceInterface := k.cli.Core.Services(k.cli.Namespace)
	for _, s := range services {
		_, err := serviceInterface.Get(s.Name, metav1.GetOptions{})
		if err == nil {
			_, err := serviceInterface.Update(s)
			if err != nil {
				return err
			}
		} else {
			_, err := serviceInterface.Create(s)
			if err != nil {
				return err
			}
		}
	}

	return k.store.Upsert(common.DefaultAppsKey,
		config.AppsVersionResource{Name: common.DefaultAppsKey, Value: apps})
}

func toConfigMap(config *models.Configuration) (*v1.ConfigMap, error) {
	configMap := &v1.ConfigMap{}
	err := copier.Copy(configMap, config)
	if err != nil {
		return nil, err
	}
	return configMap, nil
}

func toDeployAndService(app *models.Application) (*appv1.Deployment, []*v1.Service, error) {
	deploy := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
		Spec: appv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"baetyl": app.Name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{
					"baetyl": app.Name,
				}},
			},
		},
	}

	err := copier.Copy(&deploy.Spec.Template.Spec.Containers, &app.Services)
	if err != nil {
		return nil, nil, err
	}
	err = copier.Copy(&deploy.Spec.Template.Spec.Volumes, &app.Volumes)
	if err != nil {
		return nil, nil, err
	}

	var services []*v1.Service
	for _, svc := range app.Services {
		if len(svc.Ports) == 0 {
			continue
		}
		service := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      svc.Name,
				Namespace: app.Namespace,
			},
			Spec: v1.ServiceSpec{
				Selector: map[string]string{
					"baetyl": app.Name,
				},
				ClusterIP: "None",
			},
		}
		for _, p := range svc.Ports {
			port := v1.ServicePort{
				Port:       p.ContainerPort,
				TargetPort: intstr.IntOrString{IntVal: p.ContainerPort},
			}
			service.Spec.Ports = append(service.Spec.Ports, port)
		}
		services = append(services, service)
	}
	return deploy, services, nil
}
