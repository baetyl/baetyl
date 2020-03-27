package initialize

import (
	"fmt"
	"github.com/baetyl/baetyl-core/config"
	mc "github.com/baetyl/baetyl-core/mock"
	v1 "github.com/baetyl/baetyl-go/spec/v1"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	collectorBadCases = []struct {
		name         string
		fingerprints []config.Fingerprint
		err          error
	}{
		{
			name: "0: BootID Node Error",
			fingerprints: []config.Fingerprint{
				{
					Proof: config.ProofBootID,
				},
			},
			err: ErrProofValueNotFound,
		},
		{
			name: "1: SystemUUID Node Error",
			fingerprints: []config.Fingerprint{
				{
					Proof: config.ProofSystemUUID,
				},
			},
			err: ErrProofValueNotFound,
		},
		{
			name: "2: MachineID Node Error",
			fingerprints: []config.Fingerprint{
				{
					Proof: config.ProofMachineID,
				},
			},
			err: ErrProofValueNotFound,
		},
		{
			name: "3: SN File Error",
			fingerprints: []config.Fingerprint{
				{
					Proof: config.ProofSN,
					Value: "fv.txt",
				},
			},
		},
		{
			name: "4: Default",
			fingerprints: []config.Fingerprint{
				{
					Proof: config.Proof("Error"),
				},
			},
			err: ErrProofTypeNotSupported,
		},
		{
			name: "5: HostName Node Error",
			fingerprints: []config.Fingerprint{
				{
					Proof: config.ProofHostName,
				},
			},
			err: ErrProofValueNotFound,
		},
	}
)

func TestInitialize_Activate_Err_Collector(t *testing.T) {
	inspect := v1.Report{}
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	ami := mc.NewMockAMI(mockCtl)
	ami.EXPECT().Collect().Return(inspect, nil).AnyTimes()

	c := &config.Config{}
	c.Init.Cloud.Active.Interval = 5 * time.Second
	init, err := NewInit(c, ami)
	assert.Nil(t, err)

	for _, tt := range collectorBadCases {
		t.Run(tt.name, func(t *testing.T) {
			c.Init.ActivateConfig.Fingerprints = tt.fingerprints
			_, err := init.collect()
			if tt.fingerprints[0].Proof == config.ProofSN {
				assert.NotNil(t, err)
			} else {
				assert.Equal(t, tt.err, err)
			}
		})
	}
}

func TestInitialize_Activate_Err_Ami(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	ami := mc.NewMockAMI(mockCtl)
	ami.EXPECT().Collect().Return(nil, fmt.Errorf("ami error")).AnyTimes()

	c := &config.Config{}
	c.Init.Cloud.Active.Interval = 5 * time.Second
	c.Init.ActivateConfig.Fingerprints = collectorBadCases[0].fingerprints
	init, err := NewInit(c, ami)
	assert.Nil(t, err)
	_, err = init.collect()
	assert.NotNil(t, err)
}
