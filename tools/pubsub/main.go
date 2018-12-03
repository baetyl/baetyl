package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}
	switch os.Args[1] {
	case "pub":
		doPub(os.Args[2:])
	case "sub":
		doSub(os.Args[2:])
	case "perf":
		doPerf(os.Args[2:])
	default:
		usage()
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "pubsub utility")
	fmt.Fprintln(os.Stderr, "    pub [-a tcp://127.0.0.1:1883] [-u username] [-p password] [-c clientid] [-s clean_session] -t topic")
	fmt.Fprintln(os.Stderr, "    sub [-a tcp://127.0.0.1:1883] [-u username] [-p password] [-c clientid] [-s clean_session] -t topic")
}
