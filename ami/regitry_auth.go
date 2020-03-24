package ami

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type auth struct {
	Username string
	Password string
	Auth     string
}

func getRegistryToken(server, username, password string) (string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", err
	}
	auth := types.AuthConfig{
		Username:      username,
		Password:      password,
		ServerAddress: server,
	}
	res, err := cli.RegistryLogin(context.Background(), auth)
	if err != nil {
		return "", err
	}
	return res.IdentityToken, nil
}

func GenerateRegistrySecret(name, server, username, password string) (*v1.Secret, error) {
	token, err := getRegistryToken(server, username, password)
	if err != nil {
		return nil, err
	}
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Type:       v1.SecretTypeDockerConfigJson,
	}

	serverAuth := map[string]auth{
		server: {
			Username: username,
			Password: password,
			Auth:     token,
		},
	}
	auths := map[string]interface{}{
		"auths": serverAuth,
	}
	data, err := json.Marshal(auths)
	if err != nil {
		return nil, err
	}
	var res []byte
	base64.StdEncoding.Encode(res, data)
	secret.Data[v1.DockerConfigJsonKey] = res
	return secret, nil
}
