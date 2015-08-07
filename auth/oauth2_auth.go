package auth

import (
	"fmt"

	"github.com/koofr/go-httpclient"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

type CodeCallback func(authUrl string) (code string)

type OAuth2Provider struct {
	clientID       string
	clientSecret   string
	scopes         []string
	redirectURL    string
	obtainCodeFunc CodeCallback
}

func NewOAuth2Provider(id string, secret string, scopes []string, redirectURL string, cb CodeCallback) *OAuth2Provider {
	return &OAuth2Provider{
		clientID:       id,
		clientSecret:   secret,
		scopes:         scopes,
		redirectURL:    redirectURL,
		obtainCodeFunc: cb,
	}
}

func (op *OAuth2Provider) Authenticate(c *httpclient.HTTPClient) (err error) {

	conf := &oauth2.Config{
		ClientID:     op.clientID,
		ClientSecret: op.clientSecret,
		Scopes:       op.scopes,
		RedirectURL:  op.redirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/oauth2/auth", c.BaseURL.String()),
			TokenURL: fmt.Sprintf("%s/oauth2/token", c.BaseURL.String()),
		},
	}

	ctx := context.WithValue(oauth2.NoContext, oauth2.HTTPClient, c.Client)

	code := op.obtainCodeFunc(conf.AuthCodeURL("", oauth2.AccessTypeOffline))
	token, err := conf.Exchange(ctx, code)
	if err != nil {
		return
	}

	t := &oauth2.Transport{
		Base:   c.Client.Transport,
		Source: oauth2.ReuseTokenSource(token, conf.TokenSource(ctx, token)),
	}

	c.Client.Transport = t

	return
}
