package hpfeeds

import ()

// Identity will be created for each connection to allow for authentication and
// easy pub/sub checks.
type Identity struct {
	Ident  string
	Secret string

	SubChannels []string
	PubChannels []string

	auth *Authenticator // Must have some way to authenticate identities.
}
