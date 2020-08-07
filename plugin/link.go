package plugin

type Message struct {
	URI     string            `yaml:"uri" json:"uri"`
	Header  map[string]string `yaml:"header" json:"header"`
	Content interface{}       `yaml:"content" json:"content"`
}

type Link interface {
	Receive(msg *Message) error
	Request(msg *Message) (*Message, error)
	Send(msg *Message) error
	IsAsyncSupported() bool
}
