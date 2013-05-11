package main

import (
	"fmt"
	"net"
	"encoding/binary"
)

type Hpfeeds struct {
	conn *net.TCPConn
	host string
	port int
	LocalAddr net.TCPAddr
}

type msgHeader struct {
	Length uint32
	Opcode uint8
}

func NewHpfeeds(host string, port int) Hpfeeds {
	return Hpfeeds{ host: host, port: port }
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

}

func (hp *Hpfeeds) authenticate() {
	hp.readHeader()
	
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

func (hp *Hpfeeds) readString() {

}

func main() {
	hp := NewHpfeeds("hpfriends.honeycloud.net", 10000)
	hp.Connect()
}
