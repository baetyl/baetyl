package baetyl

import (
	"bytes"
	"context"
	"crypto"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/baetyl/baetyl/logger"
	schema "github.com/baetyl/baetyl/schema/v3"
	"github.com/baetyl/baetyl/utils"
)

type reportContext struct {
	log logger.Logger
	c   http.Client
}

type reportData struct {
	schema.Stats `json:",inline"`
	OTA          map[string][]*otaRecord `json:"ota,omitempty"`
}

type otaRecord struct {
	Time  string `json:"time,omitempty"`
	Step  string `json:"step,omitempty"`
	Trace string `json:"trace,omitempty"`
	Error string `json:"error,omitempty"`
}

func (rt *runtime) runReport(ctx context.Context) error {
	rc := reportContext{
		log: rt.log.WithField("ota", "report"),
		c: http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: rt.license.pool,
				},
			},
			Timeout: rt.cfg.Manage.Report.Timeout,
		},
	}
	rc.log.Infoln("begin repeatly report")
	t := time.NewTicker(rt.cfg.Manage.Report.Interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			err := rt.doReport(ctx, &rc)
			if err != nil {
				rc.log.Warnf("report fail: %s", err.Error())
			}
		case <-ctx.Done():
			rc.log.Infoln("stop repeatly report")
			return nil
		}
	}
}

func (rt *runtime) doReport(ctx context.Context, rc *reportContext) error {
	rc.log.Infoln("begin report")
	data := reportData{
		Stats: *rt.inspect(),
		OTA:   map[string][]*otaRecord{},
	}
	payload, err := json.Marshal(&data)
	if err != nil {
		return err
	}
	edata, sk, err := rt.encryptReport(payload)
	if err != nil {
		rt.log.Warnf("encrypt report fail: %s", err.Error())
		return err
	}
	buf := bytes.NewBuffer(edata)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rt.cfg.Manage.Address, buf)
	if err != nil {
		rc.log.Warnf("report create http request fail: %s", err.Error())
		return err
	}
	req.Header.Add("x-iot-edge-sn", rt.license.serial)
	req.Header.Add("x-iot-edge-key", sk)
	req.Header.Add("x-iot-edge-clientid", rt.cfg.Manage.ClientID)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := rc.c.Do(req)
	if err != nil {
		rc.log.Warnf("report http request fail: %s", err.Error())
		return err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		rc.log.Warnf("report read response fail: %s", err.Error())
		return err
	}
	rc.log.Infoln("complete report")
	return nil
}

func (rt *runtime) encryptReport(data []byte) ([]byte, string, error) {
	aesKey := utils.NewAesKey()
	body, err := utils.AesEncrypt(data, aesKey)
	if err != nil {
		return nil, "", err
	}
	s, ok := rt.license.certChain.cert.PrivateKey.(crypto.Signer)
	if !ok {
		return nil, "", errors.New("tls: private key does not provide sign")
	}
	k, err := s.Sign(nil, aesKey, crypto.Hash(0))
	if err != nil {
		return nil, "", err
	}
	key := base64.StdEncoding.EncodeToString(k)
	body = []byte(base64.StdEncoding.EncodeToString(body))
	return body, key, nil
}
