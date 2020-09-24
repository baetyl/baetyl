package plugin

import (
	"github.com/baetyl/baetyl-go/v2/mq"
)

//go:generate mockgen -destination=../mock/plugin/mq.go -package=plugin -source=mq.go MessageQueue

type MessageQueue interface {
	mq.MessageQueue
}
