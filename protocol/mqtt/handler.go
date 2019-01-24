package mqtt

import "github.com/256dpi/gomqtt/packet"

// ProcessPublish handles publish packet
type ProcessPublish func(*packet.Publish) error

// ProcessPuback handles puback packet
type ProcessPuback func(*packet.Puback) error

// ProcessError handles error
type ProcessError func(error)

// Handler MQTT message handler interface
type Handler interface {
	ProcessPublish(*packet.Publish) error
	ProcessPuback(*packet.Puback) error
	ProcessError(error)
}

// HandlerWrapper MQTT message handler wrapper
type HandlerWrapper struct {
	onPublish ProcessPublish
	onPuback  ProcessPuback
	onError   ProcessError
}

// NewHandlerWrapper creates a new handler wrapper
func NewHandlerWrapper(onPublish ProcessPublish, onPuback ProcessPuback, onError ProcessError) *HandlerWrapper {
	return &HandlerWrapper{
		onPublish: onPublish,
		onPuback:  onPuback,
		onError:   onError,
	}
}

// ProcessPublish handles publish packet
func (h *HandlerWrapper) ProcessPublish(pkt *packet.Publish) error {
	if h.onPublish == nil {
		return nil
	}
	return h.onPublish(pkt)
}

// ProcessPuback handles puback packet
func (h *HandlerWrapper) ProcessPuback(pkt *packet.Puback) error {
	if h.onPuback == nil {
		return nil
	}
	return h.onPuback(pkt)
}

// ProcessError handles error
func (h *HandlerWrapper) ProcessError(err error) {
	if h.onError == nil {
		return
	}
	h.onError(err)
}
