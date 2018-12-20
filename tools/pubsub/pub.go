package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"

	"github.com/256dpi/gomqtt/packet"
)

func doPub(args []string) {
	fs := flag.FlagSet{}
	addr := fs.String("a", "tcp://127.0.0.1:1883", "mqtt server address")
	usr := fs.String("u", "", "username")
	pwd := fs.String("p", "", "password")
	cid := fs.String("c", "", "clientid")
	topic := fs.String("t", "", "publish to topic")
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
	r := bufio.NewReader(os.Stdin)
	var pid packet.ID = 1
	for {
		line, _, err := r.ReadLine()
		if err != nil {
			if err != io.EOF {
				fmt.Fprintln(os.Stderr, "read line fail", err)
			}
			break
		}
		pub := packet.NewPublish()
		pub.ID = pid
		pub.Dup = false
		pub.Message.Topic = *topic
		pub.Message.Payload = line
		pub.Message.QOS = packet.QOSAtLeastOnce
		pub.Message.Retain = false
		err = stream.Write(pub, false)
		if err != nil {
			fmt.Fprintln(os.Stderr, "send publish packet fail", err)
			break
		}
		err = stream.Flush()
		if err != nil {
			fmt.Fprintln(os.Stderr, "flush fail", err)
			break
		}
		fmt.Println(string(line))
		pid++
	}
}
