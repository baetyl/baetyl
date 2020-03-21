package ami

import specv1 "github.com/baetyl/baetyl-go/spec/v1"

//go:generate mockgen -destination=../mock/ami.go -package=mock github.com/baetyl/baetyl-core/ami AMI

// AMI app model interfaces
type AMI interface {
	CollectInfo() (specv1.Report, error)
	ApplyApplications(specv1.Desire) error
}
