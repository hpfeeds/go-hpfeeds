package main

import (
	"../"
	"fmt"
	"os"
	"time"
)

func main() {
	host := "hpfriends.honeycloud.net"
	port := 20000
	ident := os.Args[1]
	auth := os.Args[2]

	hp := hpfeeds.NewHpfeeds(host, port, ident, auth)
	hp.Connect()

	channel := make(chan hpfeeds.Message)
	hp.Subscribe(os.Args[3], channel)
	for {
		foo := <-channel
		fmt.Println(foo.Name, string(foo.Data))
	}
}
