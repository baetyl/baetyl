package common

import (
	"sync"

	"github.com/256dpi/gomqtt/packet"
)

// AckV2 acknowledge interface
type AckV2 interface {
	Ack()
	SID() uint64
}

// PacketIDS generates packet id by sequence id for message
type PacketIDS struct {
	min     packet.ID
	max     packet.ID
	index   map[packet.ID]AckV2
	reindex map[uint64]packet.ID
	sync.RWMutex
}

// NewPacketIDS creates a new PacketIDS
func NewPacketIDS() *PacketIDS {
	return &PacketIDS{min: 1, max: 65535, index: make(map[packet.ID]AckV2), reindex: make(map[uint64]packet.ID)}
}

// Ack acknowledges by packet id
func (p *PacketIDS) Ack(id packet.ID) bool {
	p.Lock()
	ack, ok := p.index[id]
	if ok {
		delete(p.index, id)
		delete(p.reindex, ack.SID())
		ack.Ack()
	}
	p.Unlock()
	return ok
}

// Set set acknowledge with a new packet id from sequence id
func (p *PacketIDS) Set(ack AckV2) packet.ID {
	p.Lock()
	reindex := false
	id := packet.ID(ack.SID())
	for {
		if id < p.min || id > p.max {
			id = p.max
			reindex = true
		}
		if _, ok := p.index[id]; !ok {
			break
		}
		id--
		reindex = true
	}
	if reindex {
		p.reindex[ack.SID()] = id
	}
	p.index[id] = ack
	p.Unlock()
	return id
}

// Get get packet id
func (p *PacketIDS) Get(sid uint64) (pid packet.ID) {
	p.Lock()
	defer p.Unlock()
	if v, ok := p.reindex[sid]; ok {
		pid = v
	} else {
		pid = packet.ID(sid)
	}
	if _, ok := p.index[pid]; !ok {
		pid = 0
	}
	return
}

// Size returns the size of index
func (p *PacketIDS) Size() (i int) {
	p.Lock()
	i = len(p.index)
	p.Unlock()
	return
}
