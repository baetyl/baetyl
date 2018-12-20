package main

import "fmt"

func main() {
	fmt.Println("Not implemented")
}

// import (
// 	"github.com/baidu/openedge/module"
// 	"encoding/json"
// 	"flag"
// 	"fmt"
// 	"log"
// 	"sync"
// 	"sync/atomic"
// 	"time"

// 	"github.com/docker/distribution/uuid"
// 	mqtt "github.com/eclipse/paho.mqtt.golang"
// )

// var op option
// var pubClients = make([]*client, 0)
// var pubCount = uint64(0)
// var subClients = make([]*client, 0)
// var subCount = uint64(0)
// var quit = make(chan struct{})
// var pubStart, pubEnd, subEnd time.Time
// var pubToken = make(chan *mt, 100)
// var lock sync.Mutex

// type mt struct {
// 	*message
// 	mqtt.Token
// }

// type option struct {
// 	Broker   string `json:"broker"`
// 	Username string `json:"username"`
// 	Password string `json:"password"`
// 	PubNum   uint64 `json:"pubs"`
// 	SubNum   uint64 `json:"subs"`
// 	PubQos   byte   `json:"pubq"`
// 	SubQos   byte   `json:"subq"`
// 	PubTopic string `json:"pubt"`
// 	SubTopic string `json:"subt"`
// 	Total    uint64 `json:"total"`
// 	Interval uint64 `json:"interval"`
// 	Verbose  bool   `json:"verbose"`
// 	ClientID string `json:"cid"`
// }

// type message struct {
// 	ID uint64 `json:"id"`
// }

// func newMessage(p []byte) (m message) {
// 	json.Unmarshal(p, &m)
// 	return
// }

// func (m message) bytes() (v []byte) {
// 	v, _ = json.Marshal(&m)
// 	return
// }

// type client struct {
// 	sync.Mutex
// 	mqtt.Client
// 	id    string
// 	index []uint64
// }

// func newClient(id string) *client {
// 	c := &client{id: id, index: make([]uint64, op.Total)}
// 	opts := mqtt.NewClientOptions().AddBroker(op.Broker)
// 	opts.SetConnectionLostHandler(c.onLost)
// 	opts.SetOnConnectHandler(c.onConn)
// 	opts.SetCleanSession(false)
// 	opts.SetAutoReconnect(true)
// 	opts.SetKeepAlive(time.Second * 60)
// 	opts.SetPingTimeout(time.Second * 30)
// 	opts.SetUsername(op.Username)
// 	opts.SetPassword(op.Password)
// 	opts.SetClientID(id)

// 	c.Client = mqtt.NewClient(opts)
// 	for token := c.Connect(); token.Wait() && token.Error() != nil; token = c.Connect() {
// 		fmt.Printf("[%s] Connection error: %s", c.id, token.Error())
// 		time.Sleep(time.Second)
// 	}
// 	return c
// }

// func (c *client) publishing() {
// 	for {
// 		sid := atomic.AddUint64(&pubCount, 1)
// 		if sid == 1 {
// 			lock.Lock()
// 			pubStart = time.Now()
// 			lock.Unlock()
// 		} else if sid > op.Total {
// 			if sid == op.Total+1 {
// 				lock.Lock()
// 				pubEnd = time.Now()
// 				lock.Unlock()
// 			}
// 			return
// 		}
// 		if op.Interval > 0 {
// 			time.Sleep(time.Millisecond * time.Duration(op.Interval))
// 		}
// 		message := &message{sid}
// 		// if op.Qos == 0 {
// 		// 	payload := message.bytes()
// 		// 	for token := c.Publish(op.Topic, byte(op.Qos), false, payload); token.WaitTimeout(time.Second*30) && token.Error() != nil; token = c.Publish(op.Topic, byte(op.Qos), false, payload) {
// 		// 		fmt.Printf("[%s] Publish error: %s\n", c.id, token.Error())
// 		// 	}
// 		// } else {
// 		pubToken <- &mt{message: message, Token: c.Publish(op.PubTopic, byte(op.PubQos), false, message.bytes())}
// 		// }
// 	}
// }

// func (c *client) publishAcking() {
// 	for {
// 		select {
// 		case <-quit:
// 			return
// 		case t := <-pubToken:
// 			for t.WaitTimeout(time.Second*30) && t.Error() != nil {
// 				fmt.Printf("[%d] Publish error: %s\n", t.ID, t.Error())
// 				t = &mt{message: t.message, Token: c.Publish(op.PubTopic, byte(op.PubQos), false, t.message.bytes())}
// 			}
// 		}
// 	}
// }

// func (c *client) subscribe() {
// 	if token := c.Subscribe(op.SubTopic, op.SubQos, c.handleMessage); token.Wait() && token.Error() != nil {
// 		fmt.Printf("[%s] Subscribe error, %s\n", c.id, token.Error())
// 	}
// }

// func (c *client) handleMessage(client mqtt.Client, msg mqtt.Message) {
// 	// fmt.Println(string(msg.Payload()))
// 	atomic.AddUint64(&subCount, 1)
// 	message := newMessage(msg.Payload())
// 	if message.ID == 0 || message.ID > atomic.LoadUint64(&pubCount) {
// 		fmt.Println("INVALID", message)
// 		return
// 	}
// 	// fmt.Println(message)
// 	c.Lock()
// 	c.index[message.ID-1]++
// 	c.Unlock()

// 	if subCount == op.Total*op.SubNum {
// 		lock.Lock()
// 		subEnd = time.Now()
// 		lock.Unlock()
// 		close(quit)
// 	}
// }

// func (c *client) onLost(client mqtt.Client, err error) {
// 	fmt.Printf("[%s] Disconnected: %s\n", c.id, err.Error())
// }

// func (c *client) onConn(client mqtt.Client) {
// 	fmt.Printf("[%s] Connected\n", c.id)
// }

// func main() {
// 	broker := flag.String("broker", "tcp://localhost:1883", "URI of mqtt broker (required)")
// 	username := flag.String("username", "test", "Username which is used to connect to the mqtt broker")
// 	password := flag.String("password", "hahaha", "Password which is used to connect to the mqtt broker")
// 	pubnum := flag.Uint64("pubs", 1, "Number of publisher")
// 	subnum := flag.Uint64("subs", 1, "Number of subscription")
// 	pubqos := flag.Uint64("pubq", 0, "QoS level of publisher")
// 	subqos := flag.Uint64("subq", 0, "QoS level of subscription")
// 	pubtopic := flag.String("pubt", "benchmark", "Topic of publisher")
// 	subtopic := flag.String("subt", "benchmark", "Topic of subscription")
// 	total := flag.Uint64("total", 1000, "Total number of messages")
// 	interval := flag.Uint64("interval", 0, "Interval of each publish")
// 	verbose := flag.Bool("verbose", false, "Verbose")
// 	clientid := flag.String("cid", "", "Client ID")

// 	flag.Parse()

// 	op.Broker = *broker
// 	op.Username = *username
// 	op.Password = *password
// 	op.PubNum = *pubnum
// 	op.SubNum = *subnum
// 	op.PubQos = byte(*pubqos)
// 	op.SubQos = byte(*subqos)
// 	op.PubTopic = *pubtopic
// 	op.SubTopic = *subtopic
// 	op.Total = *total
// 	op.Interval = *interval
// 	op.Verbose = *verbose
// 	op.ClientID = *clientid

// 	cf, _ := json.Marshal(op)
// 	fmt.Println()
// 	fmt.Printf("Options:")
// 	fmt.Println()
// 	fmt.Printf("        %s", string(cf))
// 	fmt.Println()

// 	for i := uint64(0); i < op.SubNum; i++ {
// 		cid := op.ClientID
// 		if cid == "" {
// 			cid = uuid.Generate().String()
// 		}
// 		c := newClient(fmt.Sprintf("sub-%d-%s", i, cid))
// 		if c == nil {
// 			return
// 		}
// 		subClients = append(subClients, c)
// 	}

// 	for i := uint64(0); i < op.PubNum; i++ {
// 		cid := op.ClientID
// 		if cid == "" {
// 			cid = uuid.Generate().String()
// 		}
// 		c := newClient(fmt.Sprintf("pub-%d-%s", i, cid))
// 		if c == nil {
// 			return
// 		}
// 		pubClients = append(pubClients, c)
// 	}

// 	defer func() {
// 		for _, c := range pubClients {
// 			c.Disconnect(10)
// 		}
// 		for _, c := range subClients {
// 			c.Disconnect(10)
// 		}
// 	}()

// 	go watching()
// 	go watching2()

// 	for _, c := range subClients {
// 		c.subscribe()
// 	}

// 	fmt.Println("Publishing")
// 	for _, c := range pubClients {
// 		go c.publishing()
// 		go c.publishAcking()
// 	}

// 	<-quit
// 	time.Sleep(time.Second)

// 	lock.Lock()
// 	if pubEnd.IsZero() {
// 		pubEnd = time.Now()
// 	}
// 	pubElapsed := pubEnd.Sub(pubStart)
// 	subElapsed := subEnd.Sub(pubStart)
// 	lock.Unlock()
// 	fmt.Println()
// 	fmt.Println("Publish", pubCount-uint64(op.PubNum), "elapsed:", pubElapsed, "MPS:", float64(op.Total)/pubElapsed.Seconds())
// 	fmt.Println("Subscribe", atomic.LoadUint64(&subCount), "elapsed:", subElapsed, "MPS:", float64(subCount)/subElapsed.Seconds())
// 	fmt.Println("Validating")
// 	lostCount := 0
// 	resendCount := 0
// 	for _, c := range subClients {
// 		c.Lock()
// 		defer c.Unlock()
// 		for i, v := range c.index {
// 			if v < 1 {
// 				lostCount++
// 				if op.Verbose {
// 					fmt.Printf("[%s] Lost index: %d\n", c.id, i+1)
// 				}
// 			} else if v > 1 {
// 				resendCount++
// 			}
// 		}
// 	}
// 	fmt.Println("Lost count: ", lostCount)
// 	fmt.Println("Resend count: ", resendCount)
// 	fmt.Println("Validated")

// 	module.Wait()
// }

// func watching() {
// 	var st, et time.Time
// 	var sc, ec uint64
// 	timer := time.NewTicker(time.Second * 3)
// 	for {
// 		select {
// 		case <-timer.C:
// 			if et.IsZero() {
// 				ec = atomic.LoadUint64(&pubCount)
// 				et = time.Now()
// 				continue
// 			}
// 			sc = ec
// 			st = et
// 			ec = atomic.LoadUint64(&pubCount)
// 			et = time.Now()
// 			dt := et.Sub(st)
// 			dc := ec - sc
// 			log.Printf("Published: %d, MPS: %d, acking: %d", ec-uint64(op.PubNum), uint64(float64(dc)/dt.Seconds()), len(pubToken))
// 		case <-quit:
// 			if et.IsZero() {
// 				ec = atomic.LoadUint64(&pubCount)
// 				et = time.Now()
// 				continue
// 			}
// 			sc = ec
// 			st = et
// 			ec = atomic.LoadUint64(&pubCount)
// 			et = time.Now()
// 			dt := et.Sub(st)
// 			dc := ec - sc
// 			log.Printf("Published: %d, MPS: %d, acking: %d", ec-uint64(op.PubNum), uint64(float64(dc)/dt.Seconds()), len(pubToken))
// 			return
// 		}
// 	}
// }

// func watching2() {
// 	var st, et time.Time
// 	var sc, ec uint64
// 	timer := time.NewTicker(time.Second * 3)
// 	for {
// 		select {
// 		case <-timer.C:
// 			if et.IsZero() {
// 				ec = atomic.LoadUint64(&subCount)
// 				et = time.Now()
// 				continue
// 			}
// 			sc = ec
// 			st = et
// 			ec = atomic.LoadUint64(&subCount)
// 			et = time.Now()
// 			dt := et.Sub(st)
// 			dc := ec - sc
// 			log.Printf("Subcribed: %d, MPS: %d", ec, uint64(float64(dc)/dt.Seconds()))
// 		case <-quit:
// 			if et.IsZero() {
// 				ec = atomic.LoadUint64(&subCount)
// 				et = time.Now()
// 				continue
// 			}
// 			sc = ec
// 			st = et
// 			ec = atomic.LoadUint64(&subCount)
// 			et = time.Now()
// 			dt := et.Sub(st)
// 			dc := ec - sc
// 			log.Printf("Subcribed: %d, MPS: %d", ec, uint64(float64(dc)/dt.Seconds()))
// 			return
// 		}
// 	}
// }
