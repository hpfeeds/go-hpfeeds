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
	hp.authenticate()
	fmt.Println("Connected!")
	go hp.readLoop()
}

func (hp *Hpfeeds) Close() {
	hp.Close()
}

func (hp *Hpfeeds) authenticate() {
	hdr := hp.readHeader()

	if hdr.Opcode != OPCODE_INFO {
		panic(fmt.Sprintln("Unexpected opcode", hdr.Opcode))
	}

	name := hp.readString(0)
	nonce := make([]byte, int(hdr.Length)-(4+1)-(1+len(name)))
	hp.readData(&nonce)
	hp.sendMsgAuth(nonce)
}

func (hp *Hpfeeds) readLoop() {
	for {
		hdr := hp.readHeader()
		
		fmt.Println("hdr =", hdr)
		// break on error
		switch hdr.Opcode {
		case OPCODE_ERR:
			hp.handleError(hdr)
		case OPCODE_PUB:
			hp.handlePub(hdr)
		default:
			hp.handleUnknown(hdr)
		}
	}
}

func (hp *Hpfeeds) handleError(hdr msgHeader) {
	fmt.Println(hp.readString(uint8(hdr.Length - 5)))
}

func (hp *Hpfeeds) handlePub(hdr msgHeader) {
	name := hp.readString(0)
	channel := hp.readString(0)
	payload := hp.readString(uint8(int(hdr.Length) - (4+1) - (1+len(name)) - (1+len(channel))))
	fmt.Println("pub:", name, channel, payload)
}

func (hp *Hpfeeds) handleUnknown(hdr msgHeader) {
	fmt.Println("Unknown message type in header", hdr)
}

func (hp *Hpfeeds) readData(data interface{}) {
	err := binary.Read(hp.conn, binary.BigEndian, data)
	if err != nil {
		panic(err)
	}
}

func (hp *Hpfeeds) readHeader() msgHeader {
	hdr := msgHeader{}
	hp.readData(&hdr)
	return hdr
}

func (hp *Hpfeeds) readString(length uint8) string {

	if length == 0 {
		hp.readData(&length)
	}

	if length == 0 {
		// return err
	}

	buf := make([]byte, length)
	hp.readData(&buf)
	return string(buf)
}

func (hp *Hpfeeds) sendMsg(opcode uint8, payload []byte) {
	binary.Write(hp.conn, binary.BigEndian, msgHeader{uint32(5 + len(payload)), opcode})
//	fmt.Println("sendMsg()", payload)
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

func main() {
	hp := NewHpfeeds("hpfriends.honeycloud.net", 20000, os.Args[1], os.Args[2])
	hp.Connect()
	hp.sendMsgSub(os.Args[3])

	for {
		time.Sleep(time.Second)
	}
}
