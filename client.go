package koofrclient

import (
	"fmt"
	"github.com/koofr/go-httpclient"
	"net/http"
	"net/url"
)

type KoofrClient struct {
	*httpclient.HTTPClient
	token string
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

	return &KoofrClient{httpClient, ""}
}

func (c *KoofrClient) SetUserAgent(ua string) {
	c.Headers.Set("User-Agent", ua)
}

func (c *KoofrClient) SetToken(token string) {
	c.token = token
	c.HTTPClient.Headers.Set("Authorization", fmt.Sprintf("Token token=%s", token))
}

func (c *KoofrClient) GetToken() string {
	return c.token
}

func (c *KoofrClient) Authenticate(email string, password string) (err error) {
	var tokenResponse Token

	tokenRequest := TokenRequest{
		Email:    email,
		Password: password,
	}

	request := httpclient.RequestData{
		Method:         "POST",
		Path:           "/token",
		Headers:        make(http.Header),
		ExpectedStatus: []int{http.StatusOK},
		ReqEncoding:    httpclient.EncodingJSON,
		ReqValue:       tokenRequest,
		RespEncoding:   httpclient.EncodingJSON,
		RespValue:      &tokenResponse,
	}

	_, err = c.Request(&request)

	if err != nil {
		return
	}

	c.SetToken(tokenResponse.Token)

	return
}
