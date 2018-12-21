package client

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/256dpi/gomqtt/client/future"
	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/session"
	"github.com/256dpi/gomqtt/topic"

	"github.com/jpillora/backoff"
	"gopkg.in/tomb.v2"
)

type command struct {
	publish     bool
	subscribe   bool
	unsubscribe bool

	future        *future.Future
	message       *packet.Message
	subscriptions []packet.Subscription
	topics        []string
}

// An OnlineCallback is a function that is called when the service is connected.
//
// Note: Execution of the service is resumed after the callback returns. This
// means that waiting on a future inside the callback will deadlock the service.
type OnlineCallback func(resumed bool)

// A MessageCallback is a function that is called when a message is received.
// If an error is returned the underlying client will be prevented from
// acknowledging the specified message and closes immediately.
//
// Note: Execution of the service is resumed after the callback returns. This
// means that waiting on a future inside the callback will deadlock the service.
type MessageCallback func(*packet.Message) error

// An ErrorCallback is a function that is called when an error occurred.
//
// Note: Execution of the service is resumed after the callback returns. This
// means that waiting on a future inside the callback will deadlock the service.
type ErrorCallback func(error)

// An OfflineCallback is a function that is called when the service is disconnected.
//
// Note: Execution of the service is resumed after the callback returns. This
// means that waiting on a future inside the callback will deadlock the service.
type OfflineCallback func()

const (
	serviceStarted uint32 = iota
	serviceStopped
)

// Service is an abstraction for Client that provides a stable interface to the
// application, while it automatically connects and reconnects clients in the
// background. Errors are not returned but emitted using the ErrorCallback.
// All methods return Futures that get completed once the acknowledgements are
// received. Once the services is stopped all waiting futures get canceled.
//
// Note: If clean session is false and there are packets in the store, messages
// might get completed after starting without triggering any futures to complete.
type Service struct {
	state uint32

	config *Config

	// The session used by the client to store unacknowledged packets.
	Session Session

	// The callback that is used to notify that the service is online.
	OnlineCallback OnlineCallback

	// The callback to be called by the service upon receiving a message.
	MessageCallback MessageCallback

	// The callback to be called by the service upon encountering an error.
	ErrorCallback ErrorCallback

	// The callback that is used to notify that the service is offline.
	OfflineCallback OfflineCallback

	// The logger that is used to log write low level information like packets
	// that have ben successfully sent and received, details about the
	// automatic keep alive handler, reconnection and occurring errors.
	Logger Logger

	// The minimum delay between reconnects.
	//
	// Note: The value must be changed before calling Start.
	MinReconnectDelay time.Duration

	// The maximum delay between reconnects.
	//
	// Note: The value must be changed before calling Start.
	MaxReconnectDelay time.Duration

	// The allowed timeout until a connection attempt is canceled.
	ConnectTimeout time.Duration

	// The allowed timeout until a connection is forcefully closed.
	DisconnectTimeout time.Duration

	// The allowed timeout until a subscribe action is forcefully closed during
	// reconnect.
	ResubscribeTimeout time.Duration

	// Whether to resubscribe all subscriptions after reconnecting. Can be
	// disabled if the broker supports persistent sessions and the client is
	// configured to request one.
	ResubscribeAllSubscriptions bool

	backoff       *backoff.Backoff
	subscriptions *topic.Tree
	commandQueue  chan *command
	futureStore   *future.Store

	mutex sync.Mutex
	tomb  *tomb.Tomb
}

// NewService allocates and returns a new service. The optional parameter queueSize
// specifies how many Subscribe, Unsubscribe and Publish commands can be queued
// up before actually sending them on the wire. The default queueSize is 100.
func NewService(queueSize ...int) *Service {
	var qs = 100
	if len(queueSize) > 0 {
		qs = queueSize[0]
	}

	return &Service{
		state:                       serviceStopped,
		Session:                     session.NewMemorySession(),
		MinReconnectDelay:           50 * time.Millisecond,
		MaxReconnectDelay:           10 * time.Second,
		ConnectTimeout:              5 * time.Second,
		DisconnectTimeout:           10 * time.Second,
		ResubscribeTimeout:          5 * time.Second,
		ResubscribeAllSubscriptions: true,
		subscriptions:               topic.NewTree(),
		commandQueue:                make(chan *command, qs),
		futureStore:                 future.NewStore(),
	}
}

// Start will start the service with the specified configuration. From now on
// the service will automatically reconnect on any error until Stop is called.
func (s *Service) Start(config *Config) {
	if config == nil {
		panic("no config specified")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// return if already started
	if atomic.LoadUint32(&s.state) == serviceStarted {
		return
	}

	// set state
	atomic.StoreUint32(&s.state, serviceStarted)

	// save config
	s.config = config

	// initialize backoff
	s.backoff = &backoff.Backoff{
		Min:    s.MinReconnectDelay,
		Max:    s.MaxReconnectDelay,
		Factor: 2,
	}

	// mark future store as protected
	s.futureStore.Protect(true)

	// create new tomb
	s.tomb = new(tomb.Tomb)

	// start supervisor
	s.tomb.Go(s.supervisor)
}

// Publish will send a Publish packet containing the passed parameters. It will
// return a PublishFuture that gets completed once the quality of service flow
// has been completed.
func (s *Service) Publish(topic string, payload []byte, qos packet.QOS, retain bool) GenericFuture {
	msg := &packet.Message{
		Topic:   topic,
		Payload: payload,
		QOS:     qos,
		Retain:  retain,
	}

	return s.PublishMessage(msg)
}

// PublishMessage will send a Publish packet containing the passed message. It will
// return a PublishFuture that gets completed once the quality of service flow
// has been completed.
func (s *Service) PublishMessage(msg *packet.Message) GenericFuture {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// allocate future
	f := future.New()

	// queue publish
	s.commandQueue <- &command{
		publish: true,
		future:  f,
		message: msg,
	}

	return f
}

// Subscribe will send a Subscribe packet containing one topic to subscribe. It
// will return a SubscribeFuture that gets completed once the acknowledgements
// have been received.
func (s *Service) Subscribe(topic string, qos packet.QOS) SubscribeFuture {
	return s.SubscribeMultiple([]packet.Subscription{
		{Topic: topic, QOS: qos},
	})
}

// SubscribeMultiple will send a Subscribe packet containing multiple topics to
// subscribe. It will return a SubscribeFuture that gets completed once the
// acknowledgements have been received.
func (s *Service) SubscribeMultiple(subscriptions []packet.Subscription) SubscribeFuture {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// save subscription
	for _, v := range subscriptions {
		s.subscriptions.Set(v.Topic, v)
	}

	// allocate future
	f := future.New()

	// queue subscribe
	s.commandQueue <- &command{
		subscribe:     true,
		future:        f,
		subscriptions: subscriptions,
	}

	return &subscribeFuture{f}
}

// Unsubscribe will send a Unsubscribe packet containing one topic to unsubscribe.
// It will return a SubscribeFuture that gets completed once the acknowledgements
// have been received.
func (s *Service) Unsubscribe(topic string) GenericFuture {
	return s.UnsubscribeMultiple([]string{topic})
}

// UnsubscribeMultiple will send a Unsubscribe packet containing multiple
// topics to unsubscribe. It will return a SubscribeFuture that gets completed
// once the acknowledgements have been received.
func (s *Service) UnsubscribeMultiple(topics []string) GenericFuture {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// remove subscription
	for _, v := range topics {
		s.subscriptions.Empty(v)
	}

	// allocate future
	f := future.New()

	// queue unsubscribe
	s.commandQueue <- &command{
		unsubscribe: true,
		future:      f,
		topics:      topics,
	}

	return f
}

// Stop will disconnect the client if online and cancel all futures if requested.
// After the service is stopped in can be started again.
//
// Note: You should clear the futures on the last stop before exiting to ensure
// that all goroutines return that wait on futures.
func (s *Service) Stop(clearFutures bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// return if service not started
	if atomic.LoadUint32(&s.state) != serviceStarted {
		return
	}

	// kill and wait
	s.tomb.Kill(nil)
	s.tomb.Wait()

	// clear futures if requested
	if clearFutures {
		s.futureStore.Protect(false)
		s.futureStore.Clear()
	}

	// set state
	atomic.StoreUint32(&s.state, serviceStopped)
}

// the supervised reconnect loop
func (s *Service) supervisor() error {
	first := true

	for {
		if first {
			// no delay on first attempt
			first = false
		} else {
			// get backoff duration
			d := s.backoff.Duration()
			s.log(fmt.Sprintf("Delay Reconnect: %v", d))

			// sleep but return on Stop
			select {
			case <-time.After(d):
			case <-s.tomb.Dying():
				return tomb.ErrDying
			}
		}

		s.log("Next Reconnect")

		// prepare the stop channel
		fail := make(chan struct{})

		// try once to get a client
		client, resumed := s.connect(fail)
		if client == nil {
			continue
		}

		// resubscribe
		if s.ResubscribeAllSubscriptions {
			if !s.resubscribe(client) {
				continue
			}
		}

		// run callback
		if s.OnlineCallback != nil {
			s.OnlineCallback(resumed)
		}

		// run dispatcher on client
		dying := s.dispatcher(client, fail)

		// run callback
		if s.OfflineCallback != nil {
			s.OfflineCallback()
		}

		// return goroutine if dying
		if dying {
			return tomb.ErrDying
		}
	}
}

// will try to connect one client to the broker
func (s *Service) connect(fail chan struct{}) (*Client, bool) {
	// prepare new client
	client := New()
	client.Session = s.Session
	client.Logger = s.Logger
	client.futureStore = s.futureStore

	// set callback
	client.Callback = func(msg *packet.Message, err error) error {
		if err != nil {
			s.err("Client", err)
			close(fail)
			return nil
		}

		// call the handler
		if s.MessageCallback != nil {
			return s.MessageCallback(msg)
		}

		return nil
	}

	// attempt to connect
	connectFuture, err := client.Connect(s.config)
	if err != nil {
		s.err("Connect", err)
		return nil, false
	}

	// wait for connack
	err = connectFuture.Wait(s.ConnectTimeout)

	// check if future has been canceled
	if err == future.ErrCanceled {
		s.err("Connect", err)
		return nil, false
	}

	// check if future has timed out
	if err == future.ErrTimeout {
		client.Close()

		s.err("Connect", err)
		return nil, false
	}

	return client, connectFuture.SessionPresent()
}

func (s *Service) resubscribe(client *Client) bool {
	// get all subscriptions and return if empty
	items := s.subscriptions.All()
	if len(items) == 0 {
		return true
	}

	// prepare subscriptions
	subs := make([]packet.Subscription, 0, len(items))
	for _, v := range items {
		subs = append(subs, v.(packet.Subscription))
	}

	// sort subscriptions
	sort.Slice(subs, func(i, j int) bool {
		return subs[i].Topic < subs[j].Topic
	})

	// resubscribe all subscriptions
	subscribeFuture, err := client.SubscribeMultiple(subs)
	if err != nil {
		s.err("Resubscribe", err)
		return false
	}

	// wait for suback.
	err = subscribeFuture.Wait(s.ResubscribeTimeout)

	// check if future has been canceled
	if err == future.ErrCanceled {
		s.err("Resubscribe", err)
		return false
	}

	// check if future has timed out
	if err == future.ErrTimeout {
		client.Close()

		s.err("Resubscribe", err)
		return false
	}

	return true
}

// reads from the queues and calls the current client
func (s *Service) dispatcher(client *Client, fail chan struct{}) bool {
	for {
		select {
		case cmd := <-s.commandQueue:

			// handle subscribe command
			if cmd.subscribe {
				f2, err := client.SubscribeMultiple(cmd.subscriptions)
				if err != nil {
					s.err("Subscribe", err)

					// cancel future
					cmd.future.Cancel()

					return false
				}

				// bind future in a own goroutine. the goroutine will be
				// ultimately collected when the service is stopped
				go cmd.future.Bind(f2.(*subscribeFuture).Future)
			}

			// handle unsubscribe command
			if cmd.unsubscribe {
				f2, err := client.UnsubscribeMultiple(cmd.topics)
				if err != nil {
					s.err("Unsubscribe", err)

					// cancel future
					cmd.future.Cancel()

					return false
				}

				// bind future in a own goroutine. the goroutine will be
				// ultimately collected when the service is stopped
				go cmd.future.Bind(f2.(*future.Future))
			}

			// handle publish command
			if cmd.publish {
				f2, err := client.PublishMessage(cmd.message)
				if err != nil {
					s.err("Publish", err)

					// cancel future
					cmd.future.Cancel()

					return false
				}

				// bind future in a own goroutine. the goroutine will be
				// ultimately collected when the service is stopped
				go cmd.future.Bind(f2.(*future.Future))
			}
		case <-s.tomb.Dying():
			// disconnect client on Stop
			err := client.Disconnect(s.DisconnectTimeout)
			if err != nil {
				s.err("Disconnect", err)
			}

			return true
		case <-fail:
			return false
		}
	}
}

func (s *Service) err(sys string, err error) {
	s.log(fmt.Sprintf("%s Error: %s", sys, err.Error()))

	if s.ErrorCallback != nil {
		s.ErrorCallback(err)
	}
}

func (s *Service) log(str string) {
	if s.Logger != nil {
		s.Logger(str)
	}
}
