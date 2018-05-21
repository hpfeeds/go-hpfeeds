package main

import (
	"fmt"
	"time"

	".."
)

func main() {
	fmt.Println("Sends \"test_data\" once every second for 100 seconds.")

	host := "127.0.0.1"
	port := 10000
	ident := "test_ident"
	auth := "test_secret"

	hp := hpfeeds.NewClient(host, port, ident, auth)
	hp.Log = true
	hp.Connect()

	// Publish something on "flotest" every second
	channel1 := make(chan []byte)
	hp.Publish("test_channel", channel1)
	go func() {
		for {
			fmt.Println("Sending test_data")
			channel1 <- []byte("test_data")
			time.Sleep(time.Second)
		}
	}()
	time.Sleep(100 * time.Second)
}
