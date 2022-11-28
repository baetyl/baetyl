package initz

import (
	"encoding/json"
	"fmt"
	gohttp "net/http"
	"os"
	"path"
	"testing"
	"time"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/mock"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/baetyl/baetyl/v2/ami/kube"

	"github.com/baetyl/baetyl/v2/ami"
	"github.com/baetyl/baetyl/v2/config"
	mc "github.com/baetyl/baetyl/v2/mock"
	"github.com/baetyl/baetyl/v2/store"
)

var (
	resp = &v1.ActiveResponse{
		NodeName:  "node.test",
		Namespace: "default",
		Certificate: utils.Certificate{
			CA:                 "ca info",
			Key:                "key info",
			Cert:               "cert info",
			Name:               "name info",
			InsecureSkipVerify: false,
		},
		MqttCert: utils.Certificate{
			CA:                 "ca info",
			Key:                "key info",
			Cert:               "cert info",
			Name:               "name info",
			InsecureSkipVerify: true,
		},
	}

	goodCases = []struct {
		name         string
		fingerprints []config.Fingerprint
		want         *v1.ActiveResponse
	}{
		{
			name: "0: Pass Input",
			fingerprints: []config.Fingerprint{
				{
					Proof: config.ProofInput,
					Value: "abc",
				},
			},
			want: resp,
		},
		{
			name: "1: Pass BootID",
			fingerprints: []config.Fingerprint{
				{
					Proof: config.ProofBootID,
				},
			},
			want: resp,
		},
		{
			name: "2: Pass SystemUUID",
			fingerprints: []config.Fingerprint{
				{
					Proof: config.ProofSystemUUID,
				},
			},
			want: resp,
		},
		{
			name: "3: Pass MachineID",
			fingerprints: []config.Fingerprint{
				{
					Proof: config.ProofMachineID,
				},
			},
			want: resp,
		},
		{
			name: "4: Pass SN",
			fingerprints: []config.Fingerprint{
				{
					Proof: config.ProofSN,
					Value: "fv.txt",
				},
			},
			want: resp,
		},
		{
			name: "5: Pass HostName",
			fingerprints: []config.Fingerprint{
				{
					Proof: config.ProofHostName,
				},
			},
			want: resp,
		},
	}
)

func genActivate(t *testing.T, cfg *config.Config, ami ami.AMI) *Activate {
	ops, err := cfg.Init.Active.ToClientOptions()
	assert.NoError(t, err)
	active := &Activate{
		cfg:   cfg,
		ami:   ami,
		sig:   make(chan bool, 1),
		http:  http.NewClient(ops),
		attrs: map[string]string{},
		log:   log.With(log.Any("core", "Activate")),
	}
	active.batch = &batch{
		name:         cfg.Init.Batch.Name,
		namespace:    cfg.Init.Batch.Namespace,
		securityType: cfg.Init.Batch.SecurityType,
		securityKey:  cfg.Init.Batch.SecurityKey,
	}
	for _, a := range cfg.Init.Active.Collector.Attributes {
		active.attrs[a.Name] = a.Value
	}
	return active
}

func TestActivate(t *testing.T) {
	data, err := json.Marshal(resp)
	assert.NoError(t, err)

	r := []*mock.Response{}
	for i := 0; i < len(goodCases)+1; i++ {
		r = append(r, mock.NewResponse(200, data))
	}

	ms := mock.NewServer(nil, r...)
	assert.NotNil(t, ms)
	defer ms.Close()

	ic := &config.InitConfig{}
	err = utils.UnmarshalYAML(nil, ic)
	assert.NoError(t, err)
	ic.Active.Interval = 5 * time.Second
	ic.Batch.Name = "batch.test"
	ic.Batch.Namespace = "default"
	ic.Batch.SecurityType = "Token"
	ic.Batch.SecurityKey = "123456"
	ic.Active.Address = ms.URL
	ic.Active.Collector.Attributes = []config.Attribute{
		{
			Name:  "abc",
			Value: "abc",
		},
	}

	certPath := t.TempDir()
	var cert utils.Certificate
	err = utils.UnmarshalYAML(nil, &cert)
	assert.NoError(t, err)
	cert.Key = path.Join(certPath, "client.key")
	cert.Cert = path.Join(certPath, "client.pem")
	cert.CA = path.Join(certPath, "ca.pem")
	err = os.MkdirAll(certPath, 0755)
	assert.Nil(t, err)

	c := &config.Config{}
	c.Init = *ic
	c.Node = cert

	nodeInfo := &v1.NodeInfo{
		Hostname:         "docker-desktop",
		Address:          "192.168.1.77",
		Arch:             "amd64",
		KernelVersion:    "4.19.76-linuxkit",
		OS:               "linux",
		ContainerRuntime: "docker://19.3.5",
		MachineID:        "b49d5b1b-1c0a-42a9-9ee5-5cf69f9f8070",
		BootID:           "76a0634a-23c7-4c97-aecd-64f2b02cb267",
		SystemUUID:       "16ac43e0-0000-0000-9230-395ecd46631c",
		OSImage:          "Docker Desktop",
	}

	f, err := os.CreateTemp(t.TempDir(), t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())

	sto, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, sto)

	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	ami := mc.NewMockAMI(mockCtl)
	t.Setenv(kube.KubeNodeName, "knn")
	t.Setenv(context.KeyRunMode, context.RunModeKube)
	ami.EXPECT().CollectNodeInfo().Return(map[string]interface{}{"knn": nodeInfo}, nil).Times(len(goodCases) + 1)

	err = os.MkdirAll(defaultSNPath, 0755)
	assert.Nil(t, err)
	err = os.WriteFile(path.Join(defaultSNPath, "fv.txt"), []byte("e8fcf2c874ee46b99d2057500f6a6504"), 0755)
	assert.Nil(t, err)
	defer os.RemoveAll(path.Dir(defaultSNPath))

	for _, tt := range goodCases {
		t.Run(tt.name, func(t *testing.T) {
			c.Init.Active.Collector.Fingerprints = tt.fingerprints
			active := genActivate(t, c, ami)
			active.Start()
			active.WaitAndClose()
			responseEqual(t, *tt.want, c.Node)
		})
	}

	// mqtt link
	tt := struct {
		name         string
		fingerprints []config.Fingerprint
		want         *v1.ActiveResponse
	}{
		name: "6: Pass HostName with mqtt link",
		fingerprints: []config.Fingerprint{
			{
				Proof: config.ProofHostName,
			},
		},
		want: resp,
	}
	t.Run(tt.name, func(t *testing.T) {
		mqttCertPath := t.TempDir()
		var mqttCert utils.Certificate
		err = utils.UnmarshalYAML(nil, &mqttCert)
		assert.NoError(t, err)
		mqttCert.Key = path.Join(mqttCertPath, "client.key")
		mqttCert.Cert = path.Join(mqttCertPath, "client.pem")
		mqttCert.CA = path.Join(mqttCertPath, "ca.pem")
		err = os.MkdirAll(mqttCertPath, 0755)
		assert.Nil(t, err)

		c.Init.Active.Collector.Fingerprints = tt.fingerprints
		active := genActivate(t, c, ami)
		active.cfg.Plugin.Link = LinkMqtt
		active.cfg.MqttLink.Cert = mqttCert
		active.Start()
		active.WaitAndClose()
		responseEqual(t, *tt.want, c.Node)
	})
}

func TestActivate_Err_Response(t *testing.T) {
	errResp := map[string]string{
		"code": "ErrParam",
		"msg":  "error msg",
	}
	data, err := json.Marshal(errResp)
	assert.NoError(t, err)

	r := []*mock.Response{mock.NewResponse(500, data)}

	ms := mock.NewServer(nil, r...)
	assert.NotNil(t, ms)
	defer ms.Close()

	ic := &config.InitConfig{}
	err = utils.UnmarshalYAML(nil, ic)
	assert.NoError(t, err)
	ic.Active.Interval = 5 * time.Second
	ic.Active.Address = ms.URL
	ic.Active.Collector.Fingerprints = []config.Fingerprint{{
		Proof: config.ProofHostName,
	}}
	ic.Active.Collector.Attributes = []config.Attribute{}

	c := &config.Config{}
	c.Init = *ic

	nodeInfo := &v1.NodeInfo{
		Hostname: "docker-desktop",
	}

	f, err := os.CreateTemp("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())

	sto, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, sto)

	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	ami := mc.NewMockAMI(mockCtl)
	t.Setenv(kube.KubeNodeName, "knn")
	ami.EXPECT().CollectNodeInfo().Return(map[string]interface{}{"knn": nodeInfo}, nil).AnyTimes()

	active := genActivate(t, c, ami)
	active.Start()
	active.srv = &gohttp.Server{}
	active.Close()
}

func responseEqual(t *testing.T, resp v1.ActiveResponse, sc utils.Certificate) {
	cert, err := os.ReadFile(sc.Cert)
	assert.Nil(t, err)
	assert.Equal(t, resp.Certificate.Cert, string(cert))
	ca, err := os.ReadFile(sc.CA)
	assert.Nil(t, err)
	assert.Equal(t, resp.Certificate.CA, string(ca))
	key, err := os.ReadFile(sc.Key)
	assert.Nil(t, err)
	assert.Equal(t, resp.Certificate.Key, string(key))
}
