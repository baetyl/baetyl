package ami

import "github.com/baetyl/baetyl-go/spec/api"

// AMI app model interfaces
type AMI interface {
	CollectInfo() (*api.ReportRequest, error)
	ApplyApplications(*api.ReportResponse) error
}
