package ami

import (
	"fmt"
	"github.com/baetyl/baetyl-go/spec/crd"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"strings"
)

const (
	storageClass         = "local-storage"
	nodeAffinityLabelKey = "kubernetes.io/hostname"
)

func (k *kubeImpl) processDevices(container *v1.Container, devices []crd.Device) ([]v1.Volume, error) {
	var volumes []v1.Volume
	var volumeDevices []v1.VolumeDevice
	for _, d := range devices {
		pvcName, err := k.prepareStorage(d)
		if err != nil {
			return nil, err
		}
		device := v1.VolumeDevice{
			Name:       pvcName,
			DevicePath: d.DevicePath,
		}
		volumeDevices = append(volumeDevices, device)
		v := v1.Volume{
			Name: pvcName,
			VolumeSource: v1.VolumeSource{PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
				ClaimName: pvcName,
			}},
		}
		if d.Policy == "write" {
			v.VolumeSource.PersistentVolumeClaim.ReadOnly = false
		} else {
			v.VolumeSource.PersistentVolumeClaim.ReadOnly = true
		}
		volumes = append(volumes, v)
	}
	container.VolumeDevices = volumeDevices
	return volumes, nil
}

func (k *kubeImpl) prepareStorage(device crd.Device) (string, error) {
	node, err := k.cli.Core.Nodes().Get(k.knn, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	hostname := node.Labels[nodeAffinityLabelKey]
	if hostname == "" {
		return "", fmt.Errorf("can not get node hostname")
	}
	accessMode := v1.ReadWriteOnce
	volumeMode := v1.PersistentVolumeBlock
	storageClass := storageClass
	a := strings.Split(strings.TrimLeft(device.DevicePath, string(os.PathSeparator)), string(os.PathSeparator))
	suffix := strings.Join(a, "-")
	pvName := "device-pv-" + suffix
	quan, err := resource.ParseQuantity("1Gi")
	if err != nil {
		return "", err
	}
	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{Name: pvName},
		Spec: v1.PersistentVolumeSpec{
			StorageClassName: storageClass,
			Capacity: v1.ResourceList{
				"storage": quan,
			},
			VolumeMode:  &volumeMode,
			AccessModes: []v1.PersistentVolumeAccessMode{accessMode},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				Local: &v1.LocalVolumeSource{
					Path: device.DevicePath,
				},
			},
			PersistentVolumeReclaimPolicy: v1.PersistentVolumeReclaimRetain,
			NodeAffinity: &v1.VolumeNodeAffinity{
				Required: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{{
					MatchExpressions: []v1.NodeSelectorRequirement{{
						Key:      nodeAffinityLabelKey,
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{hostname},
					}},
				}}},
			},
		},
	}
	pvcName := "device-pvc-" + suffix
	pvc := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: pvcName, Namespace: k.cli.Namespace},
		Spec: v1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClass,
			VolumeMode:       &volumeMode,
			AccessModes:      []v1.PersistentVolumeAccessMode{accessMode},
			VolumeName:       pv.Name,
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					"storage": quan,
				},
			},
		},
	}
	_, err = k.cli.Core.PersistentVolumes().Get(pv.Name, metav1.GetOptions{})
	if err != nil {
		_, err = k.cli.Core.PersistentVolumes().Create(pv)
		if err != nil {
			return "", err
		}
	}
	_, err = k.cli.Core.PersistentVolumeClaims(k.cli.Namespace).Get(pvc.Name, metav1.GetOptions{})
	if err != nil {
		_, err = k.cli.Core.PersistentVolumeClaims(k.cli.Namespace).Create(pvc)
		if err != nil {
			return "", err
		}
	}
	return pvcName, nil
}
