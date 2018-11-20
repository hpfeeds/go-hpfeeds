package main

import (
	"fmt"

	"github.com/d1str0/go-hpfeeds"
)

func main() {
	host := "127.0.0.1"
	port := 10000
	ident := "test_ident"
	auth := "test_secret"

	hp := hpfeeds.NewClient(host, port, ident, auth)
	hp.Log = true

	msgs := make(chan hpfeeds.Message)
	go func() {
		for foo := range msgs {
			fmt.Println(foo.Name, string(foo.Payload))
		}
	}()

	for {
		fmt.Println("Connecting to hpfeeds server.")
		hp.Connect()

		// Subscribe to "flotest" and print everything coming in on it
		hp.Subscribe("test_channel", channel2)

		// Wait for disconnect
		<-hp.Disconnected
		fmt.Println("Disconnected, attemting reconnect in 10 seconds...")
		time.Sleep(10 * time.Seconds)
	}
}
