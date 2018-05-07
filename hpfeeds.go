// Package hpfeeds provides a basic client implementation of the pub/sub
// protocol. See https://github.com/rep/hpfeeds for detailed example
// descriptions of the protocol.
package hpfeeds

// These are the designated opcodes for use on the wire.
// No iota because we want to make it clear it follows the defined hpfeeds spec.
const (
	OpErr       = 0
	OpInfo      = 1
	OpAuth      = 2
	OpPublish   = 3
	OpSubscribe = 4
)

// Message describes the format of hpfeeds messages, where Name represents the
// hpfeeds identifier of the sender and Payload contains the actual data.
type Message struct {
	Name    string
	Payload []byte
}

type messageHeader struct {
	Length uint32
	Opcode uint8
}
