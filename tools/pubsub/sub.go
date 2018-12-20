package main

import (
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"

	"github.com/256dpi/gomqtt/packet"
)

func doSub(args []string) {
	fs := flag.FlagSet{}
	addr := fs.String("a", "tcp://127.0.0.1:1883", "mqtt server address")
	usr := fs.String("u", "", "username")
	pwd := fs.String("p", "", "password")
	cid := fs.String("c", "", "clientid")
	topic := fs.String("t", "", "subscribe from topic")
	err := fs.Parse(args)
	if err != nil {
		fs.Usage()
		return
	}
	if addr == nil || len(*addr) == 0 || topic == nil || len(*topic) == 0 {
		fmt.Fprintln(os.Stderr, "require address and topic")
		return
	}
	u, err := url.Parse(*addr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse address fail", err)
		return
	}
	c, err := net.Dial(u.Scheme, u.Host)
	if err != nil {
		fmt.Fprintln(os.Stderr, "connect to", *addr, "fail", err)
		return
	}
	defer c.Close()
	stream := packet.NewStream(c, c)
	conn := packet.NewConnect()
	conn.Version = packet.Version311
	conn.KeepAlive = 60
	conn.CleanSession = true
	if cid != nil {
		conn.ClientID = *cid
	}
	if usr != nil {
		conn.Username = *usr
	}
	if pwd != nil {
		conn.Password = *pwd
	}
	err = stream.Write(conn, false)
	if err != nil {
		fmt.Fprintln(os.Stderr, "send connect packet fail", err)
		return
	}
	err = stream.Flush()
	if err != nil {
		fmt.Fprintln(os.Stderr, "flush fail", err)
		return
	}
	ack, err := stream.Read()
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
	err = stream.Write(sub, false)
	if err != nil {
		fmt.Fprintln(os.Stderr, "send subscribe packet fail", err)
		return
	}
	err = stream.Flush()
	if err != nil {
		fmt.Fprintln(os.Stderr, "flush fail", err)
		return
	}
	ack, err = stream.Read()
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
	for {
		ack, err = stream.Read()
		if err != nil {
			fmt.Fprintln(os.Stderr, "read packet fail", err)
			break
		}
		pub, ok := ack.(*packet.Publish)
		if !ok {
			fmt.Fprintln(os.Stderr, "receive back packet")
			break
		}
		fmt.Println(string(pub.Message.Payload))
	}
}
