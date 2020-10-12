package pubsub

type CloudConfig struct {
	Pubsub struct {
		Size int `yaml:"size" json:"size" default:"10"`
	} `yaml:"defaultpb" json:"defaultpb"`
}
