package main

import (
	"errors"
	"log"

	"github.com/d1str0/go-hpfeeds"
)

func main() {

	db := NewDB()
	b := &hpfeeds.Broker{
		Name: "test_brkoer",
		Port: 10000,
		DB:   db,
	}
	b.SetDebugLogger(log.Print)
	b.SetInfoLogger(log.Print)
	b.SetErrorLogger(log.Print)
	err := b.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

type TestDB struct {
	IDs []hpfeeds.Identity
}

func NewDB() *TestDB {
	i := hpfeeds.Identity{
		Ident:       "test_ident",
		Secret:      "test_secret",
		SubChannels: []string{"test_channel"},
		PubChannels: []string{"test_channel"},
	}
	t := &TestDB{IDs: []hpfeeds.Identity{i}}
	return t
}

func (t *TestDB) Identify(ident string) (*hpfeeds.Identity, error) {
	if ident == "test_ident" {
		return &t.IDs[0], nil
	}
	return nil, errors.New("identifier: Unknown identity")
}
