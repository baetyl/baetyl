package trans

// Certificate certificate config for mqtt server
type Certificate struct {
	CA       string `yaml:"ca" json:"ca"`
	Key      string `yaml:"key" json:"key"`
	Cert     string `yaml:"cert" json:"cert"`
	Insecure bool   `yaml:"insecure" json:"insecure"` // for client, for test purpose
}
