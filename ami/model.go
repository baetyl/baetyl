package ami

type Model interface {
	CollectInfo() (map[string]interface{}, error)
	ApplyApplications(info map[string]interface{}) error
}
