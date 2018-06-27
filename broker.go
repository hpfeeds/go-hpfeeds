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
	ErrNilDB         = errors.New("hpfeeds: DB must not be nil")
	ErrAuthFail      = errors.New("hpfeeds: Bad credentials")
	ErrPubFail       = errors.New("hpfeeds: You do not have permission to publish to this channel")
	ErrSubFail       = errors.New("hpfeeds: You do not have permission to subscribe to this channel")
	ErrNilConn       = errors.New("hpfeeds: Session Conn is nil")
	ErrInvalidPacket = errors.New("hpfeeds: Invalid packet structure")
)

// Broker contains all needed configuration for a running broker server.
type Broker struct {
	Name        string
	Port        int
	DB          Identifier
	subMutex    sync.RWMutex
	subscribers map[string][]*Session

	clientCount int
	countMutex  sync.RWMutex

	debugLog Logger
	errorLog Logger
	infoLog  Logger
}

// ListenAndServe uses a default broker and starts serving.
func ListenAndServe(name string, port int, db Identifier) error {
	// With no special config, create new Broker with default port.
	b := &Broker{Name: name, Port: port, DB: db}
	return b.ListenAndServe()
}

// ListenAndServe starts a TCP listener and begins listening for incoming connections.
func (b *Broker) ListenAndServe() error {
	// TODO: Create a debug log function to call to pretty print this.
	b.logDebug("ListenAndServe with Broker:\n")
	b.logDebug("\tb.Name: %s\n", b.Name)
	b.logDebug("\tb.Port: %s\n", b.Port)
	b.logDebug("\tb.DB: %v\n", b.DB)

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
	b.logInfof("Now serving hpfeeds on port %d\n", b.Port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		s := NewSession(conn.(*net.TCPConn))
		//TODO: Let's print the IP of the connection here. Maybe other useful info instead of just a ptr to the Conn.
		b.logDebug("New session: %v\n", s)
		go b.serveSession(s) // Kick off the session and keep listening.
	}
}

func (b *Broker) sendInfoRequest(s *Session) error {
	// First, we must send an info message requesting auth. To do so, we first
	// generate a 4 byte nonce to send to the client.
	s.Nonce = make([]byte, SizeOfNonce)
	_, err := rand.Read(s.Nonce)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	b.logDebug("Generated nonce: %x\n", s.Nonce)
	writeField(buf, []byte(b.Name))
	buf.Write(s.Nonce)
	s.sendRawMessage(OpInfo, buf.Bytes())

	return nil
}

func (b *Broker) serveSession(s *Session) {
	b.countMutex.Lock()
	b.clientCount = b.clientCount + 1
	count := b.clientCount
	b.countMutex.Unlock()
	b.logInfof("Now serving %d clients...\n", count)

	// Defer close since we're already in a goroutine and won't be forking again.
	defer s.Close()

	b.sendInfoRequest(s)

	b.recvLoop(s)
	b.countMutex.Lock()
	b.clientCount = b.clientCount - 1
	b.countMutex.Unlock()
}

func (b *Broker) recvLoop(s *Session) {
	// Prepare a buffer for reading from the wire.
	var buf []byte

	for s.Conn != nil {
		readbuf := make([]byte, 1024)

		n, err := s.Conn.Read(readbuf)
		if err != nil {
			b.logDebug("Read(): %s\n", err)
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
	b.logDebug("Parse opcode: %d\n", opcode)
	switch opcode {
	case OpErr:
		b.logError("Received error from client: %s\n", string(data))
	case OpInfo: // Unexpected if received server side.
		b.logError("Received OpInfo from client: %s\n", string(data))
	case OpAuth:
		err := b.parseAuth(s, data)
		if err != nil {
			b.logError(err.Error())
			s.sendAuthErr()
			s.Close()
		}
	case OpPublish:
		flen := len(data)
		len1 := uint8(data[0])
		// Make sure supplied length isn't actually overbounds.
		if int(1+len1) > flen {
			b.logError("Invalid length on packet.")
			return
		}
		name := string(data[1:(1 + len1)])

		len2 := uint8(data[1+len1])
		if int(1+len1+1+len2) > flen {
			b.logError("Invalid length on packet.")
			return
		}

		channel := string(data[(1 + len1 + 1):(1 + len1 + 1 + len2)])
		payload := data[1+len1+1+len2:]
		b.handlePub(s, name, channel, payload)
	case OpSubscribe:
		flen := len(data)
		len1 := uint8(data[0])
		if int(1+len1) > flen {
			b.logError("Invalid length on packet.")
			return
		}
		name := string(data[1:(1 + len1)])
		channel := string(data[(1 + len1):])
		b.handleSub(s, name, channel)

	default:
		b.logError("Received message with unknown type %d\n", opcode)
	}
}

func (b *Broker) handleSub(s *Session, name, channel string) {
	b.logDebug("handleSub")
	b.logDebug("\tAuthenticated? %b\n", s.Authenticated)
	b.logDebug("\tName: %s\n", name)
	b.logDebug("\tChannel: %s\n", channel)
	if !s.Authenticated {
		s.sendAuthErr()
		return
	}
	id := s.Identity
	subs := id.SubChannels

	b.logDebug("%v: %v", channel, subs)
	if stringInSlice(channel, subs) {
		b.subMutex.Lock()
		b.subscribers[channel] = append(b.subscribers[channel], s)
		b.subMutex.Unlock()
	} else {
		s.sendSubErr()
		return
	}

}

func (b *Broker) handlePub(s *Session, name string, channel string, payload []byte) {
	b.logDebug("handlePub")
	b.logDebug("\tAuthenticated? %b\n", s.Authenticated)
	b.logDebug("\tName: %s\n", name)
	b.logDebug("\tChannel: %s\n", channel)
	b.logDebug("\tPayload: %x\n", payload)
	if !s.Authenticated {
		s.sendAuthErr()
		return
	}
	id := s.Identity
	pubs := id.PubChannels

	if stringInSlice(channel, pubs) {
		b.sendToChannel(name, channel, payload)
	} else {
		s.sendPubErr()
		return
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
			s.Close() // Close this session as we got an error. Not worth recovering.
			b.logError("%s\n", err.Error())
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
func (b *Broker) parseAuth(s *Session, data []byte) error {
	flen := uint8(data[0])
	if int(flen+1) > len(data) {
		return ErrInvalidPacket
	}

	ident := string(data[1 : 1+flen])
	hash := data[1+flen:]
	id, err := b.DB.Identify(ident)
	if err != nil {
		return ErrAuthFail
	}

	s.Identity = id
	s.authenticate(hash)
	if !s.Authenticated {
		return ErrAuthFail
	}
	return nil
}
