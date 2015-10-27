package auth

import (
	"fmt"
	"net/http"
	"sync"
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
