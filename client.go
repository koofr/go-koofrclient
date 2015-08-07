package koofrclient

import (
	"net/url"

	"github.com/koofr/go-httpclient"
	"github.com/koofr/go-koofrclient/auth"
)

type KoofrClient struct {
	*httpclient.HTTPClient
}

func NewKoofrClient(baseUrl string, disableSecurity bool) *KoofrClient {
	var httpClient *httpclient.HTTPClient

	if disableSecurity {
		httpClient = httpclient.Insecure()
	} else {
		httpClient = httpclient.New()
	}

	apiBaseUrl, _ := url.Parse(baseUrl)

	httpClient.BaseURL = apiBaseUrl

	httpClient.Headers.Set("User-Agent", "go koofrclient")

	return &KoofrClient{httpClient}
}

func (c *KoofrClient) SetUserAgent(ua string) {
	c.Headers.Set("User-Agent", ua)
}

func (c *KoofrClient) SetToken(token string) {
	switch t := c.Client.Transport.(type) {
	case *auth.TokenTransport:
		t.SetToken(token)
	}
}

func (c *KoofrClient) GetToken() string {
	switch t := c.Client.Transport.(type) {
	case *auth.TokenTransport:
		return t.GetToken()
	}

	// if not using token transport, return empty string (cannot return error because this would break API)
	return ""
}

func (c *KoofrClient) Authenticate(email string, password string) (err error) {
	return auth.NewTokenAuthProvider(email, password).Authenticate(c.HTTPClient)
}

func (c *KoofrClient) AuthenticateWithProvider(ap auth.AuthProvider) (err error) {
	return ap.Authenticate(c.HTTPClient)
}
