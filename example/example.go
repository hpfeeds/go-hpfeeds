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

	// Publish something on "flotest" every second
	channel1 := make(chan []byte)
	hp.Publish("flotest", channel1)
	go func() {
		for {
			time.Sleep(time.Second)
			channel1 <- []byte(fmt.Sprintf("Something"))
		}
	}()

	// Subscribe to "flotest" and print everything coming in on it
	channel2 := make(chan hpfeeds.Message)
	hp.Subscribe("flotest", channel2)
	for {
		foo := <-channel2
		fmt.Println(foo.Name, string(foo.Data))
	}
}
