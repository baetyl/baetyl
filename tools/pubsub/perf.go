package main

import (
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/256dpi/gomqtt/packet"
)

func doPerf(args []string) {
	fs := flag.FlagSet{}
	address := fs.String("a", "tcp://127.0.0.1:1883", "mqtt server address")
	username := fs.String("u", "test", "username")
	password := fs.String("p", "hahaha", "password")
	count := fs.Int("c", 10, "message count")
	topic := fs.String("t", "perf", "message topic")
	err := fs.Parse(args)
	if err != nil {
		fs.Usage()
		return
	}
	u, err := url.Parse(*address)
	if err != nil {
		fs.Usage()
		return
	}
	subc, err := net.Dial(u.Scheme, u.Host)
	if err != nil {
		fmt.Fprintln(os.Stderr, "connect to", *address, "fail", err)
		return
	}
	defer subc.Close()
	pubc, err := net.Dial(u.Scheme, u.Host)
	if err != nil {
		fmt.Fprintln(os.Stderr, "connect to", *address, "fail", err)
		return
	}
	defer pubc.Close()
	suber := packet.NewStream(subc, subc)
	puber := packet.NewStream(pubc, pubc)
	subcp := packet.NewConnect()
	subcp.Version = packet.Version311
	subcp.KeepAlive = 60
	subcp.CleanSession = true
	subcp.ClientID = fmt.Sprintf("perf-test-client-id-suber-%s", *topic)
	subcp.Username = *username
	subcp.Password = *password
	pubcp := packet.NewConnect()
	pubcp.Version = packet.Version311
	pubcp.KeepAlive = 60
	pubcp.CleanSession = true
	pubcp.ClientID = fmt.Sprintf("perf-test-client-id-puber-%s", *topic)
	pubcp.Username = *username
	pubcp.Password = *password

	err = suber.Write(subcp, false)
	if err != nil {
		fmt.Fprintln(os.Stderr, "send connect packet fail", err)
		return
	}
	err = suber.Flush()
	if err != nil {
		fmt.Fprintln(os.Stderr, "flush fail", err)
		return
	}
	ack, err := suber.Read()
	if err != nil {
		fmt.Fprintln(os.Stderr, "read packet fail", err)
	}
	connack, ok := ack.(*packet.Connack)
	if !ok {
		fmt.Fprintln(os.Stderr, "receive bad packet", ack.Type().String())
		return
	}
	if connack.ReturnCode != packet.ConnectionAccepted {
		fmt.Fprintln(os.Stderr, "connect fail", connack.ReturnCode.String())
		return
	}
	sub := packet.NewSubscribe()
	sub.ID = 1
	sub.Subscriptions = []packet.Subscription{
		packet.Subscription{
			Topic: *topic,
			QOS:   packet.QOSAtMostOnce,
		},
	}
	err = suber.Write(sub, false)
	if err != nil {
		fmt.Fprintln(os.Stderr, "send subscribe packet fail", err)
		return
	}
	err = suber.Flush()
	if err != nil {
		fmt.Fprintln(os.Stderr, "flush fail", err)
		return
	}
	ack, err = suber.Read()
	if err != nil {
		fmt.Fprintln(os.Stderr, "read packet fail", err)
		return
	}
	suback, ok := ack.(*packet.Suback)
	if !ok || suback.ID != 1 {
		fmt.Fprintln(os.Stderr, "receive bad packet", ack.Type().String())
		return
	}
	if len(suback.ReturnCodes) != 1 || suback.ReturnCodes[0] != packet.QOSAtMostOnce {
		fmt.Fprintln(os.Stderr, "subscribe fail")
		return
	}
	fmt.Fprintln(os.Stderr, "subscribe successfully")

	err = puber.Write(pubcp, false)
	if err != nil {
		fmt.Fprintln(os.Stderr, "send connect packet fail", err)
		return
	}
	err = puber.Flush()
	if err != nil {
		fmt.Fprintln(os.Stderr, "flush fail", err)
		return
	}
	ack, err = puber.Read()
	if err != nil {
		fmt.Fprintln(os.Stderr, "read packet fail", err)
	}
	connack, ok = ack.(*packet.Connack)
	if !ok {
		fmt.Fprintln(os.Stderr, "receive bad packet", ack.Type().String())
		return
	}
	if connack.ReturnCode != packet.ConnectionAccepted {
		fmt.Fprintln(os.Stderr, "connect fail", connack.ReturnCode.String())
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		start := time.Now()
		for i := 0; i < *count; i++ {
			ack, err = suber.Read()
			if err != nil {
				fmt.Fprintln(os.Stderr, "read packet fail", err)
				break
			}
			_, ok := ack.(*packet.Publish)
			if !ok {
				fmt.Fprintln(os.Stderr, "receive back packet")
				break
			}
		}
		elasped := time.Since(start)
		fmt.Println(start.Format("2006-01-02 15:04:05"), "Sub", elasped)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		start := time.Now()
		var pid packet.ID = 1
		pub := packet.NewPublish()
		pub.Dup = false
		pub.Message.Topic = *topic
		pub.Message.QOS = packet.QOSAtLeastOnce
		pub.Message.Retain = false
		for i := 0; i < *count; i++ {
			pub.ID = pid
			pub.Message.Payload = []byte(fmt.Sprintf("{\"id\":%d}", pid))

			err = puber.Write(pub, false) // TODO: test sync or async
			if err != nil {
				fmt.Fprintln(os.Stderr, "send publish packet fail", err)
				break
			}
			err = puber.Flush()
			if err != nil {
				fmt.Fprintln(os.Stderr, "flush fail", err)
				break
			}
			pid++
			if pid == 0 {
				pid++
			}
		}
		elasped := time.Since(start)
		fmt.Println(start.Format("2006-01-02 15:04:05"), "Pub", elasped)
	}()
	wg.Wait()
}
