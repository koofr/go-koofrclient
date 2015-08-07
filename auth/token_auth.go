package auth

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/davecgh/go-spew/spew"
	"github.com/koofr/go-httpclient"
)

// TokenTransport is custom transport for Go http.Client providing Koofr token auth
type TokenTransport struct {
	Provider *TokenAuthProvider
	Base     http.RoundTripper
	token    string
	lck      sync.Mutex
}

// GetToken returns current token used in this transport
func (tt *TokenTransport) GetToken() string {
	return tt.token
}

// SetToken sets new token to be used on following requests
func (tt *TokenTransport) SetToken(t string) {
	tt.token = t
}

func (tt *TokenTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {

	//cReq := cloneRequest(req)

	req.Header.Set("Authorization", fmt.Sprintf("Token token=%s", tt.GetToken()))
	resp, err = tt.Base.RoundTrip(req)

	if err != nil {
		spew.Printf("Failed request: %+v\nError: %+v\n", req, err)
		return
	}

	if resp.StatusCode == http.StatusUnauthorized {
		token, tokenErr := tt.Provider.obtainToken()
		if tokenErr != nil {
			return
		}
		tt.SetToken(token)
		//cReq := cloneRequest(req)
		if req.Method == "GET" || req.Method == "HEAD" {
			req.Header.Set("Authorization", fmt.Sprintf("Token token=%s", tt.GetToken()))
			return tt.Base.RoundTrip(req)
		}
	}
	return
}

// cloneRequest returns a clone of the provided *http.Request.
// The clone is a shallow copy of the struct and its Header map.
func cloneRequest(r *http.Request) *http.Request {
	// shallow copy of the struct
	r2 := new(http.Request)
	*r2 = *r
	// deep copy of the Header
	r2.Header = make(http.Header, len(r.Header))
	for k, s := range r.Header {
		r2.Header[k] = append([]string(nil), s...)
	}
	return r2
}

// TokenAuthProvider provides required credentials to obtain token
type TokenAuthProvider struct {
	client        *http.Client
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

// Authenticate wraps provided httpclient.HTTPClient transport with TokenTransport
func (tap *TokenAuthProvider) Authenticate(c *httpclient.HTTPClient) (err error) {

	if tap.client == nil {
		tap.client = c.Client
	}
	if tap.tokenEndpoint == "" {
		tap.tokenEndpoint = fmt.Sprintf("%s/token", c.BaseURL.String())
	}

	token, err := tap.obtainToken()
	if err != nil {
		return
	}

	// wrap it
	c.Client.Transport = &TokenTransport{
		Provider: tap,
		Base:     c.Client.Transport,
		token:    token}

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
