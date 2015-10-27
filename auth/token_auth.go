package auth

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/koofr/go-httpclient"
)

// TokenAuthProvider provides required credentials to obtain token
type TokenAuthProvider struct {
	client        *httpclient.HTTPClient
	username      string
	password      string
	tokenEndpoint string
}

func NewTokenAuthProvider(username string, password string) *TokenAuthProvider {
	return &TokenAuthProvider{
		username: username,
		password: password,
	}
}

func (tap *TokenAuthProvider) SetClient(c *httpclient.HTTPClient) {
	tap.client = c
	tap.tokenEndpoint = fmt.Sprintf("%s/token", c.BaseURL.String())
}

// Authenticate wraps provided httpclient.HTTPClient transport with TokenTransport
func (tap *TokenAuthProvider) Authenticate() (err error) {

	if tap.client == nil {
		return NotInitializedErr
	}

	token, err := tap.obtainToken()
	if err != nil {
		return
	}

	var base http.RoundTripper

	switch transport := tap.client.Client.Transport.(type) {
	case *TokenTransport:
		base = transport.Base
	default:
		base = transport
	}

	// wrap it
	tap.client.Client.Transport = &TokenTransport{
		Provider: tap,
		Base:     base,
		token:    token,
	}

	return
}

var ErrInvalidStatus = fmt.Errorf("Invalid status received")

func (tap *TokenAuthProvider) obtainToken() (token string, err error) {

	t := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		tap.username,
		tap.password,
	}

	data, err := json.Marshal(&t)
	if err != nil {
		return
	}

	tokenReq, err := http.NewRequest("POST", tap.tokenEndpoint, bytes.NewReader(data))
	if err != nil {
		return
	}
	// dont forget!
	tokenReq.Header.Set("Content-Type", "application/json")

	c := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	tokenResp, err := c.Do(tokenReq)
	if err != nil {
		return
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode != http.StatusOK {
		return "", ErrInvalidStatus
	}

	tokenContainer := struct {
		Token string
	}{}

	bytes, err := ioutil.ReadAll(tokenResp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(bytes, &tokenContainer)
	if err != nil {
		return
	}

	return tokenContainer.Token, nil
}
