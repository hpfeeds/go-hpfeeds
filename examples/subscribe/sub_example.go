package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/d1str0/hpfeeds"
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

	msgs := make(chan hpfeeds.Message)
	go func() {
		for foo := range msgs {
			fmt.Println(foo.Name, string(foo.Payload))
		}
	}()

	for {
		fmt.Println("Connecting to hpfeeds server.")
		err := hp.Connect()
		if err != nil {
			log.Fatal("Error connecting to broker server.")
		}

		// Subscribe to "flotest" and print everything coming in on it
		hp.Subscribe("test_channel", msgs)

		// Wait for disconnect
		<-hp.Disconnected
		fmt.Println("Disconnected, attemting reconnect in 10 seconds...")
		time.Sleep(10 * time.Second)
	}
}
