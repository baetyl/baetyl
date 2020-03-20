package ami

import "github.com/baetyl/baetyl-go/spec/api"

//go:generate mockgen -destination=../mock/ami.go -package=mock github.com/baetyl/baetyl-core/ami AMI

// AMI app model interfaces
type AMI interface {
	CollectInfo() (*api.ReportRequest, error)
	ApplyApplications(*api.ReportResponse) error
}
