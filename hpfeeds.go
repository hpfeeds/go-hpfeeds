package hpfeeds

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"net"
)

type Message struct {
	Name string
	Data []byte
}

type rawMsgHeader struct {
	Length uint32
	Opcode uint8
}

const (
	OPCODE_ERR  = 0
	OPCODE_INFO = 1
	OPCODE_AUTH = 2
	OPCODE_PUB  = 3
	OPCODE_SUB  = 4
)

type Hpfeeds struct {
	LocalAddr net.TCPAddr

	conn  *net.TCPConn
	host  string
	port  int
	ident string
	auth  string

	authSent     chan bool
	disconnected chan bool

	channel map[string]chan Message
}

func NewHpfeeds(host string, port int, ident string, auth string) Hpfeeds {
	return Hpfeeds{
		host:  host,
		port:  port,
		ident: ident,
		auth:  auth,

		authSent:     make(chan bool),
		disconnected: make(chan bool, 1),

		channel: make(map[string]chan Message),
	}
}

func (hp *Hpfeeds) Connect() {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", hp.host, hp.port))

	if err != nil {
		panic(err)
	}

	conn, err := net.DialTCP("tcp", &hp.LocalAddr, addr)
	if err != nil {
		panic(err)
	}

	hp.conn = conn
	go hp.recvLoop()
	<-hp.authSent
	fmt.Println("Connected!")
}

func (hp *Hpfeeds) Close() {
	hp.Close()
	disconnected <- true
}

func (hp *Hpfeeds) recvLoop() {
	buf := []byte{}
	for {
		readbuf := make([]byte, 1024)
		n, err := hp.conn.Read(readbuf)

		if err != nil {
			panic(err)
		}

		buf = append(buf, readbuf[:n]...)

		for len(buf) > 5 {
			hdr := rawMsgHeader{}
			binary.Read(bytes.NewReader(buf[0:5]), binary.BigEndian, &hdr)
			if len(buf) < int(hdr.Length) {
				break
			}
			data := buf[5:]
			hp.parse(hdr.Opcode, data)
			buf = buf[int(hdr.Length):]
		}
	}
}

func (hp *Hpfeeds) parse(opcode uint8, data []byte) {
	switch opcode {
	case OPCODE_INFO:
		hp.sendAuth(data[(1 + uint8(data[0])):])
		hp.authSent <- true
	case OPCODE_ERR:
		hp.handleError(data)
	case OPCODE_PUB:
		len1 := uint8(data[0])
		name := string(data[1:(1 + len1)])
		len2 := uint8(data[1+len1])
		channel := string(data[(1 + len1 + 1):(1 + len1 + 1 + len2)])
		payload := data[1+len1+1+len2:]
		hp.handlePub(name, channel, payload)
	default:
		hp.handleUnknown(opcode, data)
	}
}

func (hp *Hpfeeds) handleError(data []byte) {
	// TODO
	fmt.Println("error", string(data))
}

func (hp *Hpfeeds) handlePub(name string, channelName string, payload []byte) {
	channel, ok := hp.channel[channelName]
	if !ok {
		fmt.Println("Channel not subscribed.")
		return
	}
	channel <- Message{name, payload}
}

func (hp *Hpfeeds) handleUnknown(opcode uint8, data []byte) {
	// TODO
	fmt.Println("Unknown message type", opcode, data)
}

func (hp *Hpfeeds) sendRawMsg(opcode uint8, data []byte) {
	binary.Write(hp.conn, binary.BigEndian, rawMsgHeader{uint32(5 + len(data)), opcode})
	hp.conn.Write(data)
}

func (hp *Hpfeeds) sendAuth(nonce []byte) {
	mac := sha1.New()
	mac.Write(nonce)
	mac.Write([]byte(hp.auth))

	buf := new(bytes.Buffer)
	hp.writeField(buf, []byte(hp.ident), true)
	buf.Write(mac.Sum(nil))
	hp.sendRawMsg(OPCODE_AUTH, buf.Bytes())
}

func (hp *Hpfeeds) writeField(buf *bytes.Buffer, data []byte, withLength bool) {
	if withLength {
		buf.WriteByte(byte(len(data)))
	}
	buf.Write(data)
}

func (hp *Hpfeeds) sendSub(channel string) {
	buf := new(bytes.Buffer)
	hp.writeField(buf, []byte(hp.ident), true)
	hp.writeField(buf, []byte(channel), false)
	hp.sendRawMsg(OPCODE_SUB, buf.Bytes())
}

func (hp *Hpfeeds) sendPub(channel string, payload []byte) {
	buf := new(bytes.Buffer)
	hp.writeField(buf, []byte(hp.ident), true)
	hp.writeField(buf, []byte(channel), true)
	hp.writeField(buf, payload, false)
	hp.sendRawMsg(OPCODE_PUB, buf.Bytes())
}

func (hp *Hpfeeds) Subscribe(channelName string, channel chan Message) {
	hp.channel[channelName] = channel
	hp.sendSub(channelName)
}

func (hp *Hpfeeds) Unsubscribe(channelName string) {
	delete(hp.channel, channelName)
}

func (hp *Hpfeeds) Publish(channelName string, channel chan []byte) {
	go func() {
		for {
			// TODO: check if channel is still open. if not, kill this goroutine
			payload := <-channel
			hp.sendPub(channelName, payload)
		}
	}()
}
