package kube

import (
	"encoding/base64"
	"encoding/json"

	"github.com/baetyl/baetyl-go/v2/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type auth struct {
	Username string
	Password string
	Auth     string
}

func (k *kubeImpl) generateRegistrySecret(ns, name, server, username, password string) (*v1.Secret, error) {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Type:       v1.SecretTypeDockerConfigJson,
	}
	serverAuth := map[string]auth{
		server: {
			Username: username,
			Password: password,
			Auth:     base64.StdEncoding.EncodeToString([]byte(username + ":" + password)),
		},
	}
	auths := map[string]interface{}{
		"auths": serverAuth,
	}
	data, err := json.Marshal(auths)
	if err != nil {
		return nil, errors.Trace(err)
	}
	secret.Data = map[string][]byte{
		v1.DockerConfigJsonKey: data,
	}
	return secret, nil
}
