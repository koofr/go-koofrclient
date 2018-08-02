package auth

import (
	"net/http"
	"io/ioutil"
	"io"
	"fmt"
	"encoding/json"
	"bytes"
)

// That is hackish as hell, use transport to modify token expires_in property

type tokenJSON struct {
	AccessToken  string         `json:"access_token"`
	TokenType    string         `json:"token_type"`
	RefreshToken string         `json:"refresh_token"`
	ExpiresIn    int 			`json:"expires_in"`
}

type OAuth2ExchangeTransport struct {
	Base http.RoundTripper
}

func (t *OAuth2ExchangeTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {

	resp, err = t.Base.RoundTrip(req)

	if err == nil && req.URL.Path == "/oauth2/token" {
		body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1<<20))
		if err != nil {
			return nil, fmt.Errorf("oauth2: cannot fetch token: %v", err)
		}

		var tj tokenJSON
		if err = json.Unmarshal(body, &tj); err != nil {
			return nil, err
		}

		if tj.ExpiresIn > 360 {
			tj.ExpiresIn = tj.ExpiresIn - 300
		}

		newTokenBody, err := json.Marshal(tj)
		if err != nil {
			return nil, err
		}

		resp.Body = ioutil.NopCloser(bytes.NewReader(newTokenBody))
	}

	return
}
