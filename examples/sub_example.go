package main

import (
	"flag"
	"fmt"

	"github.com/d1str0/go-hpfeeds"
)

func main() {
	var (
		host    string
		port    int
		ident   string
		auth    string
		channel string
	)
	flag.StringVar(&host, "host", "127.0.0.1", "target host")
	flag.IntVar(&port, "port", 10000, "hpfeeds port")
	flag.StringVar(&ident, "ident", "test_ident", "ident username")
	flag.StringVar(&auth, "secret", "test_secret", "ident secret")
	flag.StringVar(&channel, "channel", "test_channel", "channel to subscribe to")
	flag.Parse()

	hp := hpfeeds.NewClient(host, port, ident, auth)
	hp.Log = true
	hp.Connect()

	// Subscribe to "flotest" and print everything coming in on it
	channel2 := make(chan hpfeeds.Message)
	hp.Subscribe(channel, channel2)
	go func() {
		for foo := range channel2 {
			fmt.Println(string(foo.Payload))
		}
	}()

	// Wait for disconnect
	<-hp.Disconnected
}
