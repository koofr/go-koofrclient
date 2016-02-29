package auth

import (
	"fmt"
	"net/http"

	"github.com/koofr/go-httpclient"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

type CodeCallback func(authUrl string) (code string)

type OAuth2Provider struct {
	obtainCodeFunc CodeCallback
	config         *oauth2.Config
	client         *httpclient.HTTPClient
	ctx            context.Context
}

func NewOAuth2Provider(id string, secret string, scopes []string, redirectURL string, cb CodeCallback) *OAuth2Provider {
	return &OAuth2Provider{
		obtainCodeFunc: cb,
		config: &oauth2.Config{
			ClientID:     id,
			ClientSecret: secret,
			Scopes:       scopes,
			RedirectURL:  redirectURL,
		},
	}
}

func (op *OAuth2Provider) SetClient(c *httpclient.HTTPClient) {
	op.client = c
	op.config.Endpoint = oauth2.Endpoint{
		AuthURL:  fmt.Sprintf("%s/oauth2/auth", c.BaseURL.String()),
		TokenURL: fmt.Sprintf("%s/oauth2/token", c.BaseURL.String()),
	}
	op.ctx = context.WithValue(oauth2.NoContext, oauth2.HTTPClient, c.Client)

	return
}

func (op *OAuth2Provider) GetToken() (t *oauth2.Token, err error) {

	if op.client == nil {
		return nil, NotInitializedErr
	}

	switch transport := op.client.Client.Transport.(type) {
	case *oauth2.Transport:
		return transport.Source.Token()
	}

	return nil, nil
}

func (op *OAuth2Provider) SetToken(t *oauth2.Token) (err error) {
	if op.client == nil {
		return NotInitializedErr
	}

	switch transport := op.client.Client.Transport.(type) {
	case *oauth2.Transport:
		transport.Source = oauth2.ReuseTokenSource(t, op.config.TokenSource(op.ctx, t))
		return nil
	case *TokenTransport:
		return fmt.Errorf("Invalid transport type: %T", transport)
	default:
		op.client.Client.Transport = &oauth2.Transport{
			Base:   transport,
			Source: oauth2.ReuseTokenSource(t, op.config.TokenSource(op.ctx, t)),
		}
	}
	return nil
}

func (op *OAuth2Provider) GetAuthURL(state string, opts ...oauth2.AuthCodeOption) string {
	return op.config.AuthCodeURL(state, opts...)
}

func (op *OAuth2Provider) Authenticate() (err error) {
	if op.client == nil {
		return NotInitializedErr
	}

	code := op.obtainCodeFunc(op.config.AuthCodeURL("", oauth2.AccessTypeOffline))
	op.Exchange(code)
	return
}

func (op *OAuth2Provider) Exchange(code string) (err error) {
	token, err := op.config.Exchange(op.ctx, code)
	if err != nil {
		return
	}

	var base http.RoundTripper

	switch transport := op.client.Client.Transport.(type) {
	case *oauth2.Transport:
		base = transport.Base
	default:
		base = transport
	}

	op.client.Client.Transport = &oauth2.Transport{
		Base:   base,
		Source: oauth2.ReuseTokenSource(token, op.config.TokenSource(op.ctx, token)),
	}
	return
}
