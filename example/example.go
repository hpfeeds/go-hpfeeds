package main

import (
	"fmt"
	"time"

	hpfeeds "github.com/fw42/go-hpfeeds"
	"github.com/spf13/pflag"
)

func main() {
	var host = pflag.String("host", "localhost", "host of the hpfeeds broker")
	var port = pflag.Int("port", 20000, "port of the hpfeeds broker")
	var ident = pflag.String("ident", "", "hpfeeds broker ident")
	var auth = pflag.String("auth", "", "hpfeeds auth")
	var channel = pflag.String("channel", "chan1", "hpfeeds channel to use")
	pflag.Parse()

	hp := hpfeeds.NewHpfeeds(*host, *port, *ident, *auth)
	hp.Log = true
	if err := hp.Connect(); err != nil {
		panic(err)
	}

	// Publish something on "flotest" every second
	channel1 := make(chan []byte)
	hp.Publish(*channel, channel1)
	go func() {
		for {
			channel1 <- []byte("Something")
			time.Sleep(time.Second)
		}
	}()

	// Subscribe to "flotest" and print everything coming in on it
	channel2 := make(chan hpfeeds.Message)
	hp.Subscribe(*channel, channel2)
	go func() {
		for foo := range channel2 {
			fmt.Println(foo.Name, string(foo.Payload))
		}
	}()

	// Wait for disconnect
	<-hp.Disconnected
}
