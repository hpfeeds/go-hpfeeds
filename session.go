package hpfeeds

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"log"
	"net"
)

// A session keeps track of whether or not a connection session has been
// authenticated.  Also tracks the identity of the connection.
type Session struct {
	Authenticated bool
	Identity      *Identity
	Conn          *net.TCPConn
	Nonce         []byte
}

func NewSession(conn *net.TCPConn) *Session {
	return &Session{Authenticated: false, Conn: conn, Identity: nil}
}
func (s *Session) sendAuthErr() {
	log.Println("Sending auth err")
	s.sendRawMessage(OpErr, []byte(ErrAuthFail.Error()))
}

func (s *Session) sendPubErr() {
	s.sendRawMessage(OpErr, []byte(ErrPubFail.Error()))
}
func (s *Session) sendSubErr() {
	s.sendRawMessage(OpErr, []byte(ErrSubFail.Error()))
}

func (s *Session) authenticate(clientHash []byte) {
	log.Printf("clientHash: %x\n", clientHash)
	log.Printf("ID: %v\n", s.Identity)
	mac := sha1.New()
	mac.Write(s.Nonce)
	mac.Write([]byte(s.Identity.Secret))
	servHash := mac.Sum(nil)
	log.Printf("servHash: %x\n", servHash)
	if bytes.Equal(servHash, clientHash) {
		s.Authenticated = true
	}
}
func (s *Session) sendRawMessage(opcode uint8, data []byte) error {
	log.Printf("Sending raw message: \n")
	log.Printf("Opcode: %d\n", opcode)
	log.Printf("Data Len: %d\n", len(data))
	log.Printf("Data: %x\n", data)
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
			log.Printf("Write(): %s\n", err)
			s.Conn.Close()
			s.Conn = nil
			return err
		}
		buf = buf[n:]
	}
	return nil
}
