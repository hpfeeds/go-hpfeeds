package hpfeeds

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

type Hpfeeds struct {
	LocalAddr net.TCPAddr

	conn  *net.TCPConn
	host  string
	port  int
	ident string
	auth  string

	authSent chan bool

	channel map[string]chan Message
}

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

func NewHpfeeds(host string, port int, ident string, auth string) Hpfeeds {
	return Hpfeeds{
		host:  host,
		port:  port,
		ident: ident,
		auth:  auth,

		authSent: make(chan bool),
		channel:  make(map[string]chan Message),
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
			buf = buf[int(hdr.Length):]
			hp.parsePayload(hdr.Opcode, data)
		}
	}
}

func (hp *Hpfeeds) parsePayload(opcode uint8, data []byte) {
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

func (hp *Hpfeeds) sendRawMsg(opcode uint8, payload []byte) {
	binary.Write(hp.conn, binary.BigEndian, rawMsgHeader{uint32(5 + len(payload)), opcode})
	hp.conn.Write(payload)
}

func (hp *Hpfeeds) sendAuth(nonce []byte) {
	mac := sha1.New()
	mac.Write(nonce)
	io.WriteString(mac, hp.auth)

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, uint8(len(hp.ident)))
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	io.WriteString(buf, hp.ident)
	buf.Write(mac.Sum(nil))
	hp.sendRawMsg(OPCODE_AUTH, buf.Bytes())
}

func (hp *Hpfeeds) sendSub(channel string) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, uint8(len(hp.ident)))
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	io.WriteString(buf, hp.ident)
	io.WriteString(buf, channel)
	hp.sendRawMsg(OPCODE_SUB, buf.Bytes())
}

func (hp *Hpfeeds) sendPub() {
	// TODO
}

func (hp *Hpfeeds) Subscribe(channelName string, channel chan Message) {
	hp.channel[channelName] = channel
	hp.sendSub(channelName)
}

func (hp *Hpfeeds) Unsubscribe(channelName string) {
	delete(hp.channel, channelName)
}

func (hp *Hpfeeds) Publish(channelName string, channel chan []byte) {
	// TODO
}
