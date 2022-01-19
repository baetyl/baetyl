package kube

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
)

// RPCApp use http client to call baetyl app
func (k *kubeImpl) RPCApp(url string, req *specv1.RPCRequest) (*specv1.RPCResponse, error) {
	ops := http.NewClientOptions()
	ops.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	cli := http.NewClient(ops)
	k.log.Debug("rpc http start", log.Any("url", url), log.Any("method", req.Method))

	var buf []byte
	if req.Body != nil {
		buf = []byte(fmt.Sprintf("%v", req.Body))
	}
	res, err := cli.SendUrl(strings.ToUpper(req.Method), url, bytes.NewReader(buf), req.Header)
	if err != nil {
		return nil, errors.Trace(err)
	}

	response := &specv1.RPCResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
	}
	response.Body, err = http.HandleResponse(res)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return response, nil
}
