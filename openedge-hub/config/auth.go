package config

// Permission Permission
type Permission struct {
	Action  string   `yaml:"action" json:"action" validate:"regexp=^(p|s)ub$"`
	Permits []string `yaml:"permit,flow" json:"permit,flow"`
}

// Principal Principal
type Principal struct {
	Username     string       `yaml:"username" json:"username"`
	Password     string       `yaml:"password" json:"password"`
	SerialNumber string       `yaml:"serialnumber" json:"serialnumber"`
	Permissions  []Permission `yaml:"permissions" json:"permissions"`
}
