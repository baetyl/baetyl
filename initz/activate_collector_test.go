package initz

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/baetyl/baetyl/config"
	mc "github.com/baetyl/baetyl/mock"
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

func TestActivate_Err_Collector(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	c := &config.Config{}
	c.Init.Active.Interval = 5 * time.Second

	active, err := NewActivate(c)
	assert.Error(t, err)

	ami := mc.NewMockAMI(mockCtl)
	ami.EXPECT().CollectNodeInfo().Return(nil, nil).AnyTimes()
	active = genActivate(t, c, ami)

	active.Start()
	active.Close()

	for _, tt := range collectorBadCases {
		t.Run(tt.name, func(t *testing.T) {
			c.Init.Active.Collector.Fingerprints = tt.fingerprints
			_, err := active.collect()
			if tt.fingerprints[0].Proof == config.ProofSN {
				assert.NotNil(t, err)
			} else {
				assert.Equal(t, tt.err, errors.Cause(err))
			}
		})
	}
}

func TestActivate_Err_Ami(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	ami := mc.NewMockAMI(mockCtl)
	ami.EXPECT().CollectNodeInfo().Return(nil, fmt.Errorf("ami error")).AnyTimes()

	c := &config.Config{}
	c.Init.Active.Interval = 5 * time.Second
	c.Init.Active.Collector.Fingerprints = collectorBadCases[0].fingerprints
	active := genActivate(t, c, ami)
	_, err := active.collect()
	assert.NotNil(t, err)
}
