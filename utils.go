package hpfeeds

import (
	"bytes"
	"net"
)

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func writeField(buf *bytes.Buffer, data []byte) {
	buf.WriteByte(byte(len(data)))
	buf.Write(data)
}

func deleteFromSlice(a []*Session, i int) []*Session {
	a[i] = a[len(a)-1]
	a[len(a)-1] = nil
	a = a[:len(a)-1]
	return a
}

// From net/http https://golang.org/src/net/http/server.go?s=91084:91139#L3207

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(KeepAlivePeriod)
	return tc, nil
}
