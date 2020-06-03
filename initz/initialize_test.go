package initz

import (
	"encoding/json"
	"fmt"
	"github.com/baetyl/baetyl-core/ami"
	"github.com/baetyl/baetyl-core/store"
	"github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/log"
	"io/ioutil"
	gohttp "net/http"
	"os"
	"path"
	"testing"
	"time"

	"github.com/baetyl/baetyl-core/config"
	mc "github.com/baetyl/baetyl-core/mock"
	"github.com/baetyl/baetyl-go/mock"
	v1 "github.com/baetyl/baetyl-go/spec/v1"
	"github.com/baetyl/baetyl-go/utils"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
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

func genInitialize(t *testing.T, cfg *config.Config, ami ami.AMI) *Initialize {
	ops, err := cfg.Init.Cloud.HTTP.ToClientOptions()
	assert.NoError(t, err)
	init := &Initialize{
		cfg:   cfg,
		ami:   ami,
		sig:   make(chan bool, 1),
		http:  http.NewClient(ops),
		attrs: map[string]string{},
		log:   log.With(log.Any("core", "Initialize")),
	}
	init.batch = &batch{
		name:         cfg.Init.Batch.Name,
		namespace:    cfg.Init.Batch.Namespace,
		securityType: cfg.Init.Batch.SecurityType,
		securityKey:  cfg.Init.Batch.SecurityKey,
	}
	for _, a := range cfg.Init.ActivateConfig.Attributes {
		init.attrs[a.Name] = a.Value
	}
	return init
}

func TestInitialize_Activate(t *testing.T) {
	data, err := json.Marshal(resp)
	assert.NoError(t, err)

	r := []*mock.Response{}
	for i := 0; i < len(goodCases); i++ {
		r = append(r, mock.NewResponse(200, data))
	}

	ms := mock.NewServer(nil, r...)
	assert.NotNil(t, ms)
	defer ms.Close()

	ic := &config.InitConfig{}
	err = utils.UnmarshalYAML(nil, ic)
	assert.NoError(t, err)
	ic.Cloud.Active.Interval = 5 * time.Second
	ic.Batch.Name = "batch.test"
	ic.Batch.Namespace = "default"
	ic.Batch.SecurityType = "Token"
	ic.Batch.SecurityKey = "123456"
	ic.Cloud.HTTP.Address = ms.URL
	ic.ActivateConfig.Attributes = []config.Attribute{
		{
			Name:  "abc",
			Value: "abc",
		},
	}

	certPath := "var/lib/baetyl/cert"
	is := &config.SyncConfig{}
	err = utils.UnmarshalYAML(nil, is)
	assert.NoError(t, err)
	is.Cloud.HTTP.Key = path.Join(certPath, "client.key")
	is.Cloud.HTTP.Cert = path.Join(certPath, "client.pem")
	is.Cloud.HTTP.CA = path.Join(certPath, "ca.pem")
	err = os.MkdirAll(certPath, 0755)
	assert.Nil(t, err)
	defer os.RemoveAll(path.Dir(certPath))

	c := &config.Config{}
	c.Init = *ic
	c.Sync = *is

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

	f, err := ioutil.TempFile("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())

	sto, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, sto)

	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	ami := mc.NewMockAMI(mockCtl)
	ami.EXPECT().CollectNodeInfo().Return(nodeInfo, nil).Times(len(goodCases))

	err = os.MkdirAll(defaultSNPath, 0755)
	assert.Nil(t, err)
	err = ioutil.WriteFile(path.Join(defaultSNPath, "fv.txt"), []byte("e8fcf2c874ee46b99d2057500f6a6504"), 0755)
	assert.Nil(t, err)
	defer os.RemoveAll(path.Dir(defaultSNPath))

	for _, tt := range goodCases {
		t.Run(tt.name, func(t *testing.T) {
			c.Init.ActivateConfig.Fingerprints = tt.fingerprints
			init := genInitialize(t, c, ami)
			init.Start()
			init.WaitAndClose()
			responseEqual(t, *tt.want, c.Sync)
		})
	}
}

func TestInitialize_Activate_Err_Response(t *testing.T) {
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
	ic.Cloud.Active.Interval = 5 * time.Second
	ic.Cloud.HTTP.Address = ms.URL
	ic.ActivateConfig.Fingerprints = []config.Fingerprint{{
		Proof: config.ProofHostName,
	}}
	ic.ActivateConfig.Attributes = []config.Attribute{}

	c := &config.Config{}
	c.Init = *ic

	nodeInfo := &v1.NodeInfo{
		Hostname: "docker-desktop",
	}

	f, err := ioutil.TempFile("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())

	sto, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, sto)

	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	ami := mc.NewMockAMI(mockCtl)
	ami.EXPECT().CollectNodeInfo().Return(nodeInfo, nil).AnyTimes()

	init := genInitialize(t, c, ami)
	init.Start()
	init.srv = &gohttp.Server{}
	init.Close()
}

func responseEqual(t *testing.T, resp v1.ActiveResponse, sc config.SyncConfig) {
	cert, err := ioutil.ReadFile(sc.Cloud.HTTP.Cert)
	assert.Nil(t, err)
	assert.Equal(t, resp.Certificate.Cert, string(cert))
	ca, err := ioutil.ReadFile(sc.Cloud.HTTP.CA)
	assert.Nil(t, err)
	assert.Equal(t, resp.Certificate.CA, string(ca))
	key, err := ioutil.ReadFile(sc.Cloud.HTTP.Key)
	assert.Nil(t, err)
	assert.Equal(t, resp.Certificate.Key, string(key))
}
