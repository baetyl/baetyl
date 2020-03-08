package store

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type configMapDriver struct {
	impl corev1.ConfigMapInterface
}

const ConfigMapsDriverName = "ConfigMap"

func NewConfigMapDriver(impl corev1.ConfigMapInterface) Driver {
	return &configMapDriver{
		impl: impl,
	}
}

func (cfgDriver *configMapDriver) Name() string {
	return ConfigMapsDriverName
}

func (cfgDriver *configMapDriver) Create(key []byte, val []byte) error {
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: string(key),
		},
		BinaryData: map[string][]byte{
			string(key): val,
		},
	}
	_, err := cfgDriver.impl.Create(configMap)
	if err != nil {
		return err
	}
	return nil
}

func (cfgDriver *configMapDriver) Update(key []byte, val []byte) error {
	configMap, err := cfgDriver.impl.Get(string(key), metav1.GetOptions{})
	if err != nil {
		return err
	}
	configMap.BinaryData[string(key)] = val
	_, err = cfgDriver.impl.Update(configMap)
	if err != nil {
		return err
	}
	return nil
}

func (cfgDriver *configMapDriver) Delete(key []byte) error {
	return cfgDriver.impl.Delete(string(key), &metav1.DeleteOptions{})
}

func (cfgDriver *configMapDriver) Get(key []byte) ([]byte, error) {
	configMap, err := cfgDriver.impl.Get(string(key), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return configMap.BinaryData[string(key)], nil
}

func (cfgDriver *configMapDriver) List(filter func([]byte) bool) ([]byte, error) {
	return nil, nil
}

func (cfgDriver *configMapDriver) Query(labels map[string]string) ([]byte, error) {
	return nil, nil
}
