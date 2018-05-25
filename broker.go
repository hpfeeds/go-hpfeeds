package hpfeeds

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

// Configuration
const (
	KeepAlivePeriod   = 3 * time.Minute
	DefaultBrokerPort = 10000
)

// Errors
var (
	ErrNilDB    = errors.New("hpfeeds: DB must not be nil")
	ErrAuthFail = errors.New("hpfeeds: Bad credentials")
	ErrPubFail  = errors.New("hpfeeds: You do not have permission to publish to this channel")
	ErrSubFail  = errors.New("hpfeeds: You do not have permission to subscribe to this channel")
	ErrNilConn  = errors.New("hpfeeds: Session Conn is nil")
)

// Broker contains all needed configuration for a running broker server.
type Broker struct {
	Name        string
	Port        int
	DB          Identifier
	subMutex    sync.RWMutex
	subscribers map[string][]*Session
}

// ListenAndServe uses a default broker and starts serving.
func ListenAndServe(name string, port int, db Identifier) error {
	// With no special config, create new Broker with default port.
	b := &Broker{Name: name, Port: port, DB: db}
	return b.ListenAndServe()
}

// ListenAndServe starts a TCP listener and begins listening for incoming connections.
func (b *Broker) ListenAndServe() error {
	logDebug("ListenAndServe with Broker:\n")
	logDebug("\tb.Name: %s\n", b.Name)
	logDebug("\tb.Port: %s\n", b.Port)
	logDebug("\tb.DB: %v\n", b.DB)

	if b.DB == nil {
		return ErrNilDB
	}

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", b.Port))
	if err != nil {
		return err
	}

	b.subscribers = make(map[string][]*Session)

	return b.serve(ln.(*net.TCPListener))
}

func (b *Broker) serve(ln *net.TCPListener) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		s := NewSession(conn.(*net.TCPConn))
		logDebug("New session: %v\n", s)
		go b.serveSession(s)
	}
}

func (b *Broker) serveSession(s *Session) {
	defer s.Conn.Close()

	// First, we must send an info message requesting auth. To do so, we first
	// generate a 4 byte nonce to send to the client.
	s.Nonce = make([]byte, SizeOfNonce)
	logDebug("Generated new nonce...\n")
	_, err := rand.Read(s.Nonce)
	if err != nil {
		logError("Error generating nonce: %s\n", err.Error())
		s.Conn.Close()
		return
	}

	buf := new(bytes.Buffer)
	logDebug("nonce: %x\n", s.Nonce)
	writeField(buf, []byte(b.Name))
	buf.Write(s.Nonce)
	s.sendRawMessage(OpInfo, buf.Bytes())

	b.recvLoop(s)
}

func (b *Broker) recvLoop(s *Session) {
	// Prepare a buffer for reading from the wire.
	var buf []byte

	for s.Conn != nil {
		readbuf := make([]byte, 1024)

		n, err := s.Conn.Read(readbuf)
		if err != nil {
			logDebug("Read(): %s\n", err)
			return
		}

		buf = append(buf, readbuf[:n]...)

		for len(buf) > 5 {
			hdr := messageHeader{}
			hdr.Length = binary.BigEndian.Uint32(buf[0:4]) // Get the length of the message.
			hdr.Opcode = uint8(buf[4])
			// Check to see if buf holds the full message or if we need to get more data off the wire first.
			if len(buf) < int(hdr.Length) {
				break
			}
			data := buf[5:int(hdr.Length)]
			b.parse(s, hdr.Opcode, data)
			buf = buf[int(hdr.Length):]
		}
	}
}

func (b *Broker) parse(s *Session, opcode uint8, data []byte) {
	logDebug("Parse opcode: %d\n", opcode)
	switch opcode {
	case OpErr:
		logError("Received error from client: %s\n", string(data))
	case OpInfo: // Unexpected if received server side.
		logError("Received OpInfo from client: %s\n", string(data))
	case OpAuth:
		b.parseAuth(s, data)
	case OpPublish:
		len1 := uint8(data[0])
		name := string(data[1:(1 + len1)])
		len2 := uint8(data[1+len1])
		channel := string(data[(1 + len1 + 1):(1 + len1 + 1 + len2)])
		payload := data[1+len1+1+len2:]
		b.handlePub(s, name, channel, payload)
	case OpSubscribe:
		logDebug("payload: %x\n", data)
		len1 := uint8(data[0])
		logDebug("len1: %d\n", len1)
		name := string(data[1:(1 + len1)])
		logDebug("name: %s\n", name)
		channel := string(data[(1 + len1):])
		logDebug("channel: %d\n", channel)
		b.handleSub(s, name, channel)

	default:
		logError("Received message with unknown type %d\n", opcode)
	}
}

func (b *Broker) handleSub(s *Session, name, channel string) {
	logDebug("handleSub")
	logDebug("\tAuthenticated? %b\n", s.Authenticated)
	logDebug("\tName: %s\n", name)
	logDebug("\tChannel: %s\n", channel)
	if !s.Authenticated {
		s.sendAuthErr()
	}
	id := s.Identity
	subs := id.SubChannels

	logDebug("%v: %v", channel, subs)
	if stringInSlice(channel, subs) {
		b.subMutex.Lock()
		b.subscribers[channel] = append(b.subscribers[channel], s)
		b.subMutex.Unlock()
	} else {
		s.sendSubErr()
	}

}

func (b *Broker) handlePub(s *Session, name string, channel string, payload []byte) {
	log.Debug("handlePub")
	logDebug("\tAuthenticated? %b\n", s.Authenticated)
	logDebug("\tName: %s\n", name)
	logDebug("\tChannel: %s\n", channel)
	logDebug("\tPayload: %x\n", payload)
	if !s.Authenticated {
		s.sendAuthErr()
	}
	id := s.Identity
	pubs := id.PubChannels

	if stringInSlice(channel, pubs) {
		b.sendToChannel(name, channel, payload)
	} else {
		s.sendPubErr()
	}
}

func (b *Broker) sendToChannel(name string, channel string, payload []byte) {
	buf := new(bytes.Buffer)
	writeField(buf, []byte(name))
	writeField(buf, []byte(channel))
	writeField(buf, payload)

	b.subMutex.RLock()
	sessions := b.subscribers[channel]

	for _, s := range sessions {
		err := s.sendRawMessage(OpPublish, buf.Bytes())
		if err != nil {
			if s.Conn != nil {
				s.Conn.Close()
				s.Conn = nil
			}
			logError("%s\n", err.Error())
			defer b.pruneSessions(channel)
		}
	}
	b.subMutex.RUnlock()
}

// Remove any closed Sessions.
func (b *Broker) pruneSessions(channel string) {
	b.subMutex.Lock()
	defer b.subMutex.Unlock()

	var valid []*Session
	for _, s := range b.subscribers[channel] {
		if s.Conn != nil {
			valid = append(valid, s)
		}
	}
	b.subscribers[channel] = valid
}

// Parse an auth request.
func (b *Broker) parseAuth(s *Session, data []byte) {
	logDebug("Parse auth: %x\n", data)
	len := uint8(data[0])
	logDebug("len: %d\n", len)
	ident := string(data[1 : 1+len])
	logDebug("ident: %s\n", ident)
	hash := data[1+len:]
	logDebug("hash: %x\n", hash)
	id, err := b.DB.Identify(ident)
	if err != nil {
		logError("Failure identifying ident: %v\n", err)
		s.sendAuthErr()
		s.Conn.Close()
		return
	}
	s.Identity = id
	s.authenticate(hash)
	if !s.Authenticated {
		s.sendAuthErr()
	}
}
