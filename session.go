package hpfeeds

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"net"
	"sync"
)

// A session keeps track of whether or not a connection session has been
// authenticated.  Also tracks the identity of the connection.
type Session struct {
	Authenticated bool
	Identity      *Identity
	Conn          *net.TCPConn
	connMutex     *sync.RWMutex
	Nonce         []byte
}

func NewSession(conn *net.TCPConn) *Session {
	return &Session{Authenticated: false, Conn: conn, Identity: nil}
}

// Close will close the connection and set Conn to nil. This is a thread safe function.
func (s *Session) Close() {
	s.connMutex.Lock()
	if s.Conn != nil {
		s.Conn.Close()
		s.Conn = nil
	}
	s.connMutex.Unlock()
}

func (s *Session) sendAuthErr() {
	s.sendRawMessage(OpErr, []byte(ErrAuthFail.Error()))
}

func (s *Session) sendPubErr() {
	s.sendRawMessage(OpErr, []byte(ErrPubFail.Error()))
}
func (s *Session) sendSubErr() {
	s.sendRawMessage(OpErr, []byte(ErrSubFail.Error()))
}

func (s *Session) authenticate(clientHash []byte) {
	mac := sha1.New()
	mac.Write(s.Nonce)
	mac.Write([]byte(s.Identity.Secret))
	servHash := mac.Sum(nil)
	if bytes.Equal(servHash, clientHash) {
		s.Authenticated = true
	}
}

func (s *Session) sendRawMessage(opcode uint8, data []byte) error {
	s.connMutex.RLock() // Don't defer since we might unlock sooner for Close()
	if s.Conn == nil {
		return ErrNilConn
	}

	buf := make([]byte, 5)
	binary.BigEndian.PutUint32(buf, uint32(5+len(data)))
	buf[4] = byte(opcode)
	buf = append(buf, data...)
	for len(buf) > 0 {
		n, err := s.Conn.Write(buf)
		if err != nil {
			s.connMutex.RUnlock()
			s.Close()
			return err
		}
		buf = buf[n:]
	}
	s.connMutex.RUnlock()
	return nil
}
