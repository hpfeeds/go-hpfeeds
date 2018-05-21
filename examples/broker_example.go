package main

import (
	"errors"
	"log"

	".."
)

func main() {

	db := NewDB()
	err := hpfeeds.ListenAndServe("test_broker", 10000, db)
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
