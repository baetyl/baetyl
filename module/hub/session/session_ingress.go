package session

import (
	"github.com/256dpi/gomqtt/packet"
)

func (s *session) callback(pid uint32) {
	ack := packet.NewPuback()
	ack.ID = packet.ID(pid)
	err := s.send(ack, true)
	if err != nil {
		s.close(true)
		return
	}
	s.log.Debugf("Send puback(pid=%d) successfully", pid)
}
