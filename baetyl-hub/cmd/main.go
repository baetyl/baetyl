package main

import (
	hub "github.com/baetyl/baetyl/baetyl-hub"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
)

func main() {
	baetyl.Run(hub.Run)
}
