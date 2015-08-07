package auth

import "github.com/koofr/go-httpclient"

type AuthProvider interface {
	Authenticate(client *httpclient.HTTPClient) error
}
