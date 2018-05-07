package main

import (
	".."
	"fmt"
)

func main() {
	host := "54.80.116.47"
	port := 10000
	ident := "test_ident"
	auth := "test_secret"

	hp := hpfeeds.NewClient(host, port, ident, auth)
	hp.Log = true
	hp.Connect()

	// Subscribe to "flotest" and print everything coming in on it
	channel2 := make(chan hpfeeds.Message)
	hp.Subscribe("test_channel", channel2)
	go func() {
		for foo := range channel2 {
			fmt.Println(foo.Name, string(foo.Payload))
		}
	}()

	// Wait for disconnect
	<-hp.Disconnected
}
