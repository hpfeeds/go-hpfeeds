package hpfeeds

// Authenticator is an interface to abstract possible storage mechanisms for
// Auth credentials. Possibilities include MongoDB, BoltDB, or flat config
// files.
type Identifier interface {
	// Take username/password an
	Identify(ident string) (*Identity, error)
}

// Identity will be created for each connection to allow for authentication and
// easy pub/sub checks.
type Identity struct {
	Ident  string // TODO: Rename to "Name"
	Secret string

	SubChannels []string
	PubChannels []string
}
