// Package common 公共实现定义
package common

import (
	"sync"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/google/uuid"
)

const (
	RequestID   = "x-baetyl-request-id"
	FlagSyncMsg = "sync"
)

type SendKeeper struct {
	results sync.Map
}

type Send func(*v1.Message) error

func (s *SendKeeper) SendSync(msg *v1.Message, timeout time.Duration, send Send) (*v1.Message, error) {
	if msg.Metadata == nil {
		msg.Metadata = make(map[string]string)
	}
	reqID := uuid.New().String()
	msg.Metadata[RequestID] = reqID
	msg.Metadata[FlagSyncMsg] = "true"
	ch := make(chan *v1.Message, 1)
	_, ok := s.results.LoadOrStore(reqID, ch)
	if ok {
		return nil, errors.Errorf("request id collision")
	}
	defer s.results.Delete(reqID)
	if err := send(msg); err != nil {
		return nil, err
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil, errors.Errorf("request timeout")
	case msg := <-ch:
		return msg, nil
	}
}

func (s *SendKeeper) ReceiveResp(msg *v1.Message) error {
	if msg.Metadata == nil {
		return errors.Errorf("message meta data is empty")
	}
	reqID, ok := msg.Metadata[RequestID]
	if !ok {
		return errors.Errorf("failed to get request id")
	}
	val, ok := s.results.Load(reqID)
	if !ok {
		return errors.Errorf("failed to get related channel")
	}
	ch := val.(chan *v1.Message)
	select {
	case ch <- msg:
	default:
	}
	return nil
}

func IsSyncMessage(msg *v1.Message) bool {
	if msg.Metadata == nil {
		return false
	}
	return msg.Metadata[FlagSyncMsg] != ""
}
