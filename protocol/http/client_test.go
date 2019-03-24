package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_SendUrl(t *testing.T) {
	t.Skip("need network")
	var cfg ClientInfo
	cfg.CA = "../../example/native/var/db/openedge/localhub-cert-only-for-test/ca.pem"
	cli, err := NewClient(cfg)
	assert.NoError(t, err)
	url := "https://www.baidu.com/"
	res, err := cli.SendUrl("GET", url, nil, nil)
	assert.NoError(t, err)
	defer res.Close()
}
