package function

import (
	"encoding/json"
	"fmt"

	"github.com/256dpi/gomqtt/packet"
)

// MakeErrorPayload makes error payload
func MakeErrorPayload(p *packet.Publish, err error) []byte {
	s := make(map[string]interface{})
	s["packet"] = p
	s["errorMessage"] = err.Error()
	s["errorType"] = fmt.Sprintf("%T", err)
	o, _ := json.Marshal(s)
	return o
}
