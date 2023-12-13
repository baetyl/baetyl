// Package ws 实现端云基于ws协议的链接
package ws

import (
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"

	specV1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl/v2/plugin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func prepareServer(t *testing.T) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {
		u := websocket.Upgrader{}
		c, err := u.Upgrade(w, r, nil)
		assert.NoError(t, err)

		var msg specV1.Message
		err = c.ReadJSON(&msg)
		assert.NoError(t, err)

		err = c.WriteJSON(msg)
		assert.NoError(t, err)
		err = c.Close()
		assert.NoError(t, err)
	})
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	go func() {
		server.ListenAndServeTLS(
			"./testcert/server.pem", "./testcert/server.key")
	}()
	return server
}

func prepareClient(t *testing.T) *wsLink {
	tempDir, err := ioutil.TempDir("", t.Name())
	assert.NoError(t, err)
	confFile := filepath.Join(tempDir, "service.yml")
	confString := `
wslink:
  address: wss://127.0.0.1:8080
  syncUrl: sync
  insecureSkipVerify: true
node:
  ca: ./testcert/ca.pem
  key: ./testcert/client.key
  cert: ./testcert/client.pem
  insecureSkipVerify: true
`
	err = ioutil.WriteFile(confFile, []byte(confString), 0755)
	assert.NoError(t, err)
	plugin.ConfFile = confFile
	plg, err := New()
	assert.NoError(t, err)
	link, ok := plg.(*wsLink)
	assert.True(t, ok)
	return link
}

func TestSendAsyncAndReceive(t *testing.T) {
	server := prepareServer(t)
	ws := prepareClient(t)

	msg := &specV1.Message{
		Kind: specV1.MessageReport,
		Content: specV1.LazyValue{Value: specV1.Desire{
			"123": "456",
		}},
	}
	err := ws.Send(msg)
	assert.NoError(t, err)
	msgCh, errCh := ws.Receive()
	select {
	case res := <-msgCh:
		assert.Equal(t, msg.Kind, res.Kind)
		assert.EqualValues(t, msg.Metadata, res.Metadata)
		dr := specV1.Desire{}
		err = msg.Content.Unmarshal(&dr)
		assert.NoError(t, err)
		assert.EqualValues(t, msg.Content.Value, dr)
	case err = <-errCh:
		assert.NoError(t, err)
	}

	err = server.Close()
	assert.NoError(t, err)

	err = ws.Close()
	assert.NoError(t, err)
}

func TestSendSync(t *testing.T) {
	server := prepareServer(t)
	ws := prepareClient(t)

	msg := &specV1.Message{
		Kind: specV1.MessageReport,
		Content: specV1.LazyValue{Value: specV1.Desire{
			"123": "456",
		}},
	}
	res, err := ws.Request(msg)
	assert.NoError(t, err)
	assert.Equal(t, msg.Kind, res.Kind)
	assert.EqualValues(t, msg.Metadata, res.Metadata)
	dr := specV1.Desire{}
	err = msg.Content.Unmarshal(&dr)
	assert.NoError(t, err)
	assert.EqualValues(t, msg.Content.Value, dr)

	err = server.Close()
	assert.NoError(t, err)

	err = ws.Close()
	assert.NoError(t, err)
}
