package hpfeeds

// Authenticator is an interface to abstract possible storage mechanisms for
// Auth credentials. Possibilities include MongoDB, BoltDB, or flat config
// files.
type Authenticator interface {
	// Take username/password and return whether or not they are a valid account.
	Authenticate(ident, secret string) bool

	// Return the list of subscribed channels belonging to this user.
	SubChannels(ident) []string

	// Return the list of Publishable channels belonging to this user.
	PubChannels(ident) []string
}
