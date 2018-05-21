package main

import (
	".."
	"fmt"
)

func main() {
	host := "54.80.116.47"
	port := 10000
	ident := "mhn-global-collector"
	auth := "a7sd8djd8djd6dh3hd7dj0hdsjssj"

	hp := hpfeeds.NewClient(host, port, ident, auth)
	hp.Log = true
	hp.Connect()

	// Subscribe to "flotest" and print everything coming in on it
	channel2 := make(chan hpfeeds.Message)
	hp.Subscribe("mhn-community-v2.events", channel2)
	go func() {
		i := 0
		for foo := range channel2 {
			if i%10 == 0 {
				fmt.Println(foo.Name, string(foo.Payload))
				fmt.Printf("Total records: %d\n", i)
			}
			i++
		}
	}()

	// Wait for disconnect
	<-hp.Disconnected
}
