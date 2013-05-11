package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

type Hpfeeds struct {
	conn      *net.TCPConn
	host      string
	port      int
	ident     string
	auth      string
	LocalAddr net.TCPAddr
}

type msgHeader struct {
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
	return Hpfeeds{host: host, port: port, ident: ident, auth: auth}
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
	fmt.Println("Connected!")
	go hp.recvLoop()
	// wait until authenticated
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
			hdr := msgHeader{}
			binary.Read(bytes.NewReader(buf[0:5]), binary.BigEndian, &hdr)
			if len(buf) < int(hdr.Length) {
				break
			}
			data := buf[5:]
			buf = buf[int(hdr.Length):]
			hp.parseMessage(hdr.Opcode, data)
		}
	}
}

func (hp *Hpfeeds) parseMessage(opcode uint8, data []byte) {
	switch opcode {
	case OPCODE_INFO:
		hp.sendMsgAuth(data[(1 + uint8(data[0])):])
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
	fmt.Println(string(data))
}

func (hp *Hpfeeds) handlePub(name string, channel string, payload []byte) {
	fmt.Println("pub:", name, channel, string(payload))
}

func (hp *Hpfeeds) handleUnknown(opcode uint8, data []byte) {
	// TODO
	fmt.Println("Unknown message type", opcode, data)
}

func (hp *Hpfeeds) sendMsg(opcode uint8, payload []byte) {
	binary.Write(hp.conn, binary.BigEndian, msgHeader{uint32(5 + len(payload)), opcode})
	hp.conn.Write(payload)
}

func (hp *Hpfeeds) sendMsgAuth(nonce []byte) {
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
	hp.sendMsg(OPCODE_AUTH, buf.Bytes())
}

func (hp *Hpfeeds) sendMsgSub(channel string) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, uint8(len(hp.ident)))
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	io.WriteString(buf, hp.ident)
	io.WriteString(buf, channel)
	hp.sendMsg(OPCODE_SUB, buf.Bytes())
}

func (hp *Hpfeeds) sendMsgPub() {
	// TODO
}

func (hp *Hpfeeds) Subscribe() {
	// TODO
}

func (hp *Hpfeeds) Publish() {
	// TODO
}

func (hp *Hpfeeds) SubscribeJSON() {
	// TODO
}

func (hp *Hpfeds) PublishJSON() {
	// TODO
}

func main() {
	hp := NewHpfeeds("hpfriends.honeycloud.net", 20000, os.Args[1], os.Args[2])
	hp.Connect()
	time.Sleep(time.Second)
	hp.sendMsgSub(os.Args[3])

	for {
		time.Sleep(time.Second)
	}
}
