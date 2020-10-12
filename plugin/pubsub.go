package plugin

import (
	"github.com/baetyl/baetyl-go/v2/pubsub"
)

//go:generate mockgen -destination=../mock/plugin/pubsub.go -package=plugin -source=pubsub.go Pubsub

type Pubsub interface {
	pubsub.Pubsub
}
