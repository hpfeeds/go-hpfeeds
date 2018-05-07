package hpfeeds

import (
	"testing"
)

const (
	TestHost  = "hpfeeds.test.com"
	TestPort  = 10000
	TestIdent = "hpfeeds-client"
	TestAuth  = "client-secret"
)

func TestClient_NewClient(t *testing.T) {
	c := NewClient(TestHost, TestPort, TestIdent, TestAuth)
	if c.Host != TestHost {
		t.Error("Host mismatch.")
	}
	if c.Port != TestPort {
		t.Error("Port mismatch.")
	}
	if c.Ident != TestIdent {
		t.Error("Ident mismatch.")
	}
	if c.Auth != TestAuth {
		t.Error("Auth mismatch.")
	}
}
