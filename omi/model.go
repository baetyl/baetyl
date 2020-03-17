package omi

type Model interface {
	CollectInfo(map[string]string) (map[string]interface{}, error)
	ApplyApplications(info map[string]string) error
}
