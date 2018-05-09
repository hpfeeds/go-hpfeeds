package hpfeeds

import ()

const (
	ErrNilAuth = errors.New("hpfeeds: Authenticator must not be nil.")
)

type Broker struct {
	Port int
}

func ListenAndServe(auth *Authenticator) error {
	b := &Broker{Port: 10000}
	return b.ListenAndServe(auth)
}

func (b *Broker) ListenAndServe(auth *Authenticator) error {
	if auth == nil {
		return ErrNilAuth
	}

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", b.Port))
	if err != nil {
		return err
	}

	return b.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)})
}

func (b *Broker) Serve(ln *net.TCPListener) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go handleConnection(conn)
	}
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
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}
