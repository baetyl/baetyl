package agent

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"

	"github.com/baidu/openedge/module/http"
	"github.com/baidu/openedge/module/utils"
)

// Report Report
func (a *Agent) report(keyFile string, data []byte) error {
	body, key, err := encryptData(keyFile, data)
	if err != nil {
		return err
	}
	headers := http.Headers{}
	headers.Set("x-iot-edge-clientid", a.conf.ClientID)
	headers.Set("x-iot-edge-key", key)
	headers.Set("Content-Type", "application/x-www-form-urlencoded")
	_, _, err = a.http.Send("POST", fmt.Sprintf("%s://%s/v1/edge/info", a.http.Addr.Scheme, a.http.Addr.Host), headers, body)
	return err
}

// Download download config file
func (a *Agent) download(keyFile, url string) ([]byte, error) {
	reqHeaders := http.Headers{}
	reqHeaders.Set("x-iot-edge-clientid", a.conf.ClientID)
	reqHeaders.Set("Content-Type", "application/octet-stream")
	resHeaders, resBody, err := a.http.Send("GET", url, reqHeaders, nil)
	if err != nil {
		return nil, err
	}
	data, err := decryptData(keyFile, resHeaders.Get("x-iot-edge-key"), resBody)
	return data, err
}

func encryptData(keyPath string, data []byte) ([]byte, string, error) {
	priKey, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, "", err
	}
	aesKey := utils.NewAesKey()
	// encrypt data using AES
	body, err := utils.AesEncrypt(data, aesKey)
	if err != nil {
		return nil, "", err
	}
	// encrypt AES key using RSA
	k, err := utils.RsaPrivateEncrypt(aesKey, priKey)
	if err != nil {
		return nil, "", err
	}
	// encode key using BASE64
	key := base64.StdEncoding.EncodeToString(k)
	// encode body using BASE64
	body = []byte(base64.StdEncoding.EncodeToString(body))
	return body, key, nil
}

func decryptData(keyPath, key string, data []byte) ([]byte, error) {
	priKey, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	// decode key using BASE64
	k, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}
	// decrypt AES key using RSA
	aesKey, err := utils.RsaPrivateDecrypt(k, priKey)
	if err != nil {
		return nil, err
	}
	// decrypt data using AES
	decData, err := utils.AesDecrypt(data, aesKey)
	return decData, err
}
