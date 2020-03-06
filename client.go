// Package hpfeeds provides a basic implementation of the pub/sub protocol
// of the Honeynet Project. See https://github.com/rep/hpfeeds for detailed
// descriptions of the protocol
package hpfeeds

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

// Client stores internal state for on connection. On disconnection,
// the Disconnected channel (buffered) will be written. Set LocalAddr to set a
// local IP address and port which the connection should bind to on connect.
// TODO: Connect timeout?
type Client struct {
	LocalAddr net.TCPAddr

	conn  *net.TCPConn
	Host  string
	Port  int
	Ident string
	Auth  string

	authSent     chan bool
	Disconnected chan error

	channel map[string]chan Message

	Log bool
}

// NewClient returns a new Client object and initializes necessary channels.
func NewClient(host string, port int, ident string, auth string) Client {
	return Client{
		Host:  host,
		Port:  port,
		Ident: ident,
		Auth:  auth,

		authSent:     make(chan bool),
		Disconnected: make(chan error, 1),

		channel: make(map[string]chan Message),
	}
}

// Connect establishes a new hpfeeds connection and will block until the
// connection is successfully estabilshed or the connection attempt failed.
// TODO: Rename to Dial()?
func (c *Client) Connect() error {
	c.clearDisconnected()

	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port))
	if err != nil {
		return err
	}

	// Do we need two steps here?
	conn, err := net.DialTCP("tcp", &c.LocalAddr, addr)
	if err != nil {
		return err
	}

	c.conn = conn
	go c.recvLoop()
	<-c.authSent

	select {
	case err = <-c.Disconnected:
		return err
	default:
	}

	return nil
}

// TODO: Better way to do this?
// Does this block?
func (c *Client) clearDisconnected() {
	select {
	case <-c.Disconnected:
	default:
	}
}

// Returns given error on the Disconnected chan.
func (c *Client) setDisconnected(err error) {
	c.clearDisconnected()
	c.Disconnected <- err
}

// Close closes the hpfeeds connection and signals the Disconnected channel.
func (c *Client) Close() {
	c.close(nil)
}

func (c *Client) close(err error) {
	c.conn.Close()
	c.setDisconnected(err)
	select {
	case c.authSent <- false:
	default:
	}
	c.conn = nil // Do we want to set to nil?
}

func (c *Client) recvLoop() {
	// Prepare a buffer for reading from the wire.
	var buf []byte

	for c.conn != nil {
		readbuf := make([]byte, 1024)

		n, err := c.conn.Read(readbuf)
		if err != nil {
			c.log("Read(): %s\n", err)
			c.close(err)
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
			c.parse(hdr.Opcode, data)
			buf = buf[int(hdr.Length):]
		}
	}
}

func (c *Client) parse(opcode uint8, data []byte) {
	switch opcode {
	case OpInfo:
		log.Printf("Recieved OpInfo:\n")
		log.Printf("Opcode: %d\n", opcode)
		log.Printf("Data len: %d\n", len(data))
		log.Printf("data: %x\n", data)
		c.sendAuth(data[(1 + uint8(data[0])):])
		c.authSent <- true
	case OpErr:
		c.log("Received error from server: %s\n", string(data))
	case OpPublish:
		flen := len(data)
		if flen == 0 {
			c.log("Invalid packet length. Data size is 0")
			return
		}
		len1 := uint8(data[0])

		if int(1+len1) > flen {
			c.log("Invalid packet length for len of name")
			return
		}
		name := string(data[1:(1 + len1)])
		len2 := uint8(data[1+len1])

		// Expect payload of at least 1, hence +1 at end.
		if int(1+len1+1+len2+1) > flen {
			c.log("Invalid packet length for data provided")
			return
		}
		channel := string(data[(1 + len1 + 1):(1 + len1 + 1 + len2)])
		payload := data[1+len1+1+len2:]
		c.handlePub(name, channel, payload)
	default:
		c.log("Received message with unknown type %d\n", opcode)
	}
}

func (c *Client) handlePub(name string, channelName string, payload []byte) {
	channel, ok := c.channel[channelName]
	if !ok {
		c.log("Received message on unsubscribed channel %s\n", channelName)
		return
	}
	channel <- Message{name, payload}
}

func (c *Client) sendRawMsg(opcode uint8, data []byte) {
	buf := make([]byte, 5)
	binary.BigEndian.PutUint32(buf, uint32(5+len(data)))
	buf[4] = byte(opcode)
	buf = append(buf, data...)
	for len(buf) > 0 {
		n, err := c.conn.Write(buf)
		if err != nil {
			c.log("Write(): %s\n", err)
			c.close(err)
			return
		}
		buf = buf[n:]
	}
}

func (c *Client) sendAuth(nonce []byte) {
	log.Printf("nonce: %x\n", nonce)
	buf := new(bytes.Buffer)
	mac := sha1.New()
	mac.Write(nonce)
	mac.Write([]byte(c.Auth))
	writeField(buf, []byte(c.Ident))
	buf.Write(mac.Sum(nil))
	c.sendRawMsg(OpAuth, buf.Bytes())
}

func (c *Client) sendSub(channelName string) {
	buf := new(bytes.Buffer)
	writeField(buf, []byte(c.Ident))
	buf.Write([]byte(channelName))
	c.sendRawMsg(OpSubscribe, buf.Bytes())
}

func (c *Client) sendPub(channelName string, payload []byte) {
	buf := new(bytes.Buffer)
	writeField(buf, []byte(c.Ident))
	writeField(buf, []byte(channelName))
	buf.Write(payload)
	c.sendRawMsg(OpPublish, buf.Bytes())
}

// Subscribe sends a subscribe message to the hpfeeds server. All incoming
// messages on the given hpfeeds channel will now be written to the given Go
// channel.
func (c *Client) Subscribe(channelName string, channel chan Message) {
	c.channel[channelName] = channel
	c.sendSub(channelName)
}

// Publish starts a new goroutine which reads from the given Go channel
// and for each item sends a publish message to the given hpfeeds channel.
// If the Go channel is externally closed, the goroutine will exit.
func (c *Client) Publish(channelName string, channel chan []byte) {
	go func() {
		for payload := range channel {
			if c.conn == nil {
				return
			}
			c.sendPub(channelName, payload)
		}
	}()
}

func (c *Client) log(format string, v ...interface{}) {
	if c.Log {
		log.Printf(format, v...)
	}
}
