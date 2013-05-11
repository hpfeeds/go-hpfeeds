package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
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
	fmt.Println("name:", name)

	nonce := make([]byte, int(hdr.Length)-(4+1)-(1+len(name)))
	hp.readData(&nonce)
	fmt.Println("nonce:", nonce)
	hp.sendAuth(nonce, hp.ident, hp.auth)
	fmt.Println("auth done")

	hp.readLoop()
}

func (hp *Hpfeeds) readLoop() {
	for {
		hdr := hp.readHeader()
		switch hdr.Opcode {
		case OPCODE_ERR:
			fmt.Println(hp.readString(uint8(hdr.Length - 5)))
		}
	}
}

func (hp *Hpfeeds) readData(data interface{}) {
	err := binary.Read(hp.conn, binary.BigEndian, data)
	if err != nil {
		panic(err)
	}
	fmt.Println("readData(): ", data)
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
	fmt.Println("sendMsg()", payload)
	hp.conn.Write(payload)
}

func (hp *Hpfeeds) sendAuth(nonce []byte, ident string, auth string) {
	mac := sha1.New()
	mac.Write(nonce)
	io.WriteString(mac, auth)

	buf := new(bytes.Buffer)
	fmt.Println("ident =", ident)
	err := binary.Write(buf, binary.BigEndian, uint8(len(ident)))
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	io.WriteString(buf, ident)
	fmt.Println("mac =", mac.Sum(nil))
	buf.Write(mac.Sum(nil))
	hp.sendMsg(OPCODE_AUTH, buf.Bytes())
}

func main() {
	fmt.Println(os.Args)
	hp := NewHpfeeds("hpfriends.honeycloud.net", 10000, os.Args[1], os.Args[2])
	hp.Connect()
}
