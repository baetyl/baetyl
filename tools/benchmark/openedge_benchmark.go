package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/docker/distribution/uuid"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// ClientOption ClientOption
type ClientOption struct {
	Broker   string
	Action   string
	Qos      int
	Retain   bool
	Topic    string
	Username string
	Password string
	// Count        int
	ClientNum   int
	KeepAlive   int
	MessageSize int
	// IntervalTime int
}

var sendCount uint64
var receiveCount uint64

var connLostHandler MQTT.ConnectionLostHandler = func(client MQTT.Client, err error) {
	log.Printf("already disconnect, %s\n", err.Error())
}

func execute(f func(clients []MQTT.Client, opts ClientOption), opts ClientOption) {
	clients := make([]MQTT.Client, 0)
	for i := 0; i < opts.ClientNum; i++ {
		client := connect(i, opts)
		if client == nil {
			break
		}
		clients = append(clients, client)
	}

	if len(clients) != opts.ClientNum {
		for _, v := range clients {
			disConnect(v)
		}
		return
	}
	f(clients, opts)
}

func connect(id int, clientOption ClientOption) MQTT.Client {
	opts := MQTT.NewClientOptions().AddBroker(clientOption.Broker)
	opts.SetConnectionLostHandler(connLostHandler)
	opts.SetAutoReconnect(false)
	opts.SetKeepAlive(time.Second * time.Duration(clientOption.KeepAlive))
	opts.SetUsername(clientOption.Username)
	opts.SetPassword(clientOption.Password)

	if clientOption.Action == "sub" {
		opts.SetClientID(fmt.Sprintf("sub-%d-%s", id, uuid.Generate().String()))
	} else if clientOption.Action == "pub" {
		opts.SetClientID(fmt.Sprintf("pub-%d-%s", id, uuid.Generate().String()))
	} else {
		log.Println("Unrecognized action")
		return nil
	}

	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		log.Printf("Connection error, %s\n", token.Error())
		return nil
	}
	return c
}

func disConnect(client MQTT.Client) {
	client.Disconnect(10)
}

func startPublish(clients []MQTT.Client, opts ClientOption) {
	message := createFixedSizeMessage(opts.MessageSize)
	for i, c := range clients {
		if !c.IsConnected() {
			c = connect(i, opts)
		}
		go func(client MQTT.Client) {
			for {
				publish(client, opts, message)
			}
		}(c)
	}

	var st, et time.Time
	var sc, ec uint64
	timer := time.NewTicker(time.Second * 5)
	for range timer.C {
		if et.IsZero() {
			ec = atomic.LoadUint64(&sendCount)
			et = time.Now()
			continue
		}
		sc = ec
		st = et
		ec = atomic.LoadUint64(&sendCount)
		et = time.Now()
		dt := et.Sub(st)
		dc := ec - sc
		log.Printf("Sent %d message, throughput = %d message/second", ec, uint64(float64(dc)/dt.Seconds()))
	}
}

func publish(client MQTT.Client, opts ClientOption, message string) {
	if token := client.Publish(opts.Topic, byte(opts.Qos), opts.Retain, message); token.Wait() && token.Error() != nil {
		log.Printf("Publish error, %s\n", token.Error())
	} else {
		atomic.AddUint64(&sendCount, 1)
	}
}

func startSubscribe(clients []MQTT.Client, opts ClientOption) {
	for i, c := range clients {
		if !c.IsConnected() {
			c = connect(i, opts)
		}
		go subscribe(c, opts.Topic, byte(opts.Qos))
	}

	var st, et time.Time
	var sc, ec uint64
	timer := time.NewTicker(time.Second * 5)
	for range timer.C {
		if et.IsZero() {
			ec = atomic.LoadUint64(&receiveCount)
			et = time.Now()
			continue
		}
		sc = ec
		st = et
		ec = atomic.LoadUint64(&receiveCount)
		et = time.Now()
		dt := et.Sub(st)
		dc := ec - sc
		log.Printf("Received %d message, throughput = %d message/second", ec, uint64(float64(dc)/dt.Seconds()))
	}
}

func subscribe(client MQTT.Client, topic string, qos byte) {
	if token := client.Subscribe(topic, qos, func(client MQTT.Client, msg MQTT.Message) {
		atomic.AddUint64(&receiveCount, 1)
	}); token.Wait() && token.Error() != nil {
		log.Printf("Subscribe error, %s\n", token.Error())
	}
}

func createFixedSizeMessage(size int) string {
	var buffer bytes.Buffer
	for i := 0; i < size; i++ {
		buffer.WriteString(strconv.Itoa(i % 10))
	}

	message := buffer.String()
	return message
}

func main() {
	broker := flag.String("broker", "tcp://localhost:1883", "URI of MQTT broker (required)")
	qos := flag.Int("qos", 0, "MQTT QoS level 0|1")
	action := flag.String("action", "", "pub: Publish, sub: Subscribe")
	retain := flag.Bool("retain", false, "Message retain")
	topic := flag.String("topic", "benchmark", "Base topic")
	username := flag.String("username", "test", "Username which is used to connect to the MQTT broker")
	password := flag.String("password", "hahaha", "Password which is used to connect to the MQTT broker")
	// count := flag.Int("count", 1, "Number of loops per client")
	clientNum := flag.Int("clients", 10, "Number of clients")
	keepalive := flag.Int("keepAlive", 600, "Keepalive reconnect to broker (s)")
	msgSize := flag.Int("messageSize", 1024, "Message size of per publish (byte)")
	// intervalTime := flag.Int("intervalTime", 10, "Interval time per message (ms)")

	flag.Parse()

	if *action != "pub" && *action != "sub" {
		flag.Usage()
		return
	}

	var clientOption ClientOption
	clientOption.Broker = *broker
	clientOption.Qos = *qos
	clientOption.Action = *action
	clientOption.Retain = *retain
	clientOption.Topic = *topic
	clientOption.MessageSize = *msgSize
	clientOption.Username = *username
	clientOption.Password = *password
	// clientOption.IntervalTime = *intervalTime
	clientOption.ClientNum = *clientNum
	// clientOption.Count = *count
	clientOption.KeepAlive = *keepalive

	cf, _ := json.Marshal(clientOption)
	log.Println()
	log.Printf("Options:")
	log.Println()
	log.Printf("        %s", string(cf))
	log.Println()

	// Validate "action"
	if clientOption.Action == "pub" {
		execute(startPublish, clientOption)
	} else if clientOption.Action == "sub" {
		execute(startSubscribe, clientOption)
	} else {
		log.Println("Unrecognized action")
		return
	}
}
