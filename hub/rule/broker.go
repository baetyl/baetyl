package rule

import (
	"github.com/baidu/openedge/hub/common"
	"github.com/baidu/openedge/hub/config"
)

// Broker broker interface
type broker interface {
	Config() *config.Config
	MsgQ0Chan() <-chan *common.Message
	Flow(msg *common.Message)
	FetchQ1(offset uint64, batchSize int) ([]*common.Message, error)
	OffsetChanLen() int
	OffsetPersisted(id string) (*uint64, error)
	PersistOffset(id string, offset uint64) error
	InitOffset(id string, persistent bool) (uint64, error)
	WaitOffsetPersisted()
}
