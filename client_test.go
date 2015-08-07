package koofrclient_test

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"

	. "github.com/koofr/go-koofrclient"
	"github.com/koofr/go-koofrclient/auth"
	"github.com/nu7hatch/gouuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type fakeKoofr struct {
	*http.ServeMux
	tokens     map[string]time.Time
	tokensLock sync.RWMutex
}

func (k *fakeKoofr) ListTokens() (tokens []string) {
	k.tokensLock.RLock()
	defer k.tokensLock.RUnlock()
	tokens = make([]string, 0, len(k.tokens))
	for t, _ := range k.tokens {
		tokens = append(tokens, t)
	}
	return
}

func (k *fakeKoofr) InvalidateToken(t string) {
	k.tokensLock.Lock()
	delete(k.tokens, t)
	k.tokensLock.Unlock()
}

func (k *fakeKoofr) withAuth(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		t := strings.Split(r.Header.Get("Authorization"), "=")[1]

		k.tokensLock.RLock()
		_, has := k.tokens[t]
		k.tokensLock.RUnlock()
		if has == false {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			f(w, r)
		}
	}
}

func (k *fakeKoofr) issueToken(w http.ResponseWriter, r *http.Request) {
	k.tokensLock.Lock()
	defer k.tokensLock.Unlock()
	body, err := ioutil.ReadAll(r.Body)
	Expect(err).NotTo(HaveOccurred())

	tr := &struct {
		Email    string
		Password string
	}{}

	err = json.Unmarshal(body, tr)
	Expect(err).NotTo(HaveOccurred())

	if tr.Email == "" && tr.Password == "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	token, _ := uuid.NewV4()
	k.tokens[token.String()] = time.Now().Add(time.Hour)

	fmt.Fprintf(w, fmt.Sprintf(`{"token": "%s"}`, token.String()))
}

func (k *fakeKoofr) mounts(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, fmt.Sprintf(`{"mounts": []}`))
	return
}

func (k *fakeKoofr) devices(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		w.Header().Add("Content-Type", "application/json")
		fmt.Fprint(w, fmt.Sprintf(`{"devices": []}`))
	case "POST":
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, fmt.Sprintf(`{"devices": []}`))
	}
	return
}

func (k *fakeKoofr) setup() {
	k.HandleFunc("/token", k.issueToken)
	k.HandleFunc("/api/v2/mounts", k.withAuth(k.mounts))
	k.HandleFunc("/api/v2/mounts/", k.withAuth(k.mounts))
	k.HandleFunc("/api/v2/devices", k.withAuth(k.devices))
	k.HandleFunc("/api/v2/devices/", k.withAuth(k.devices))
}

func NewFakeKoofr() *fakeKoofr {
	k := &fakeKoofr{
		ServeMux: http.NewServeMux(),
		tokens:   make(map[string]time.Time),
	}

	k.setup()

	return k
}

var _ = Describe("Koofrclient", func() {
	var (
		koofr  *fakeKoofr
		client *KoofrClient
		server *httptest.Server
	)

	BeforeEach(func() {
		koofr = NewFakeKoofr()
		server = httptest.NewTLSServer(koofr)
		client = NewKoofrClient(server.URL, true)
		client.Client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
	})

	AfterEach(func() {
		server.Close()
		server.CloseClientConnections()
	})

	It("should authenticate", func() {
		err := client.Authenticate("test", "testPass")
		Expect(err).NotTo(HaveOccurred())
		Expect(client.GetToken()).NotTo(BeEmpty())
		_, err = client.Mounts()
		Expect(err).NotTo(HaveOccurred())
	})

	It("should handle 403 from server", func() {
		err := client.Authenticate("", "")
		Expect(err).To(HaveOccurred())
	})

	Context("with authentication done", func() {

		JustBeforeEach(func() {
			err := client.AuthenticateWithProvider(auth.NewTokenAuthProvider("user", "pass"))
			Expect(err).NotTo(HaveOccurred())
			Expect(client.GetToken()).NotTo(BeEmpty())
		})

		It("should renew token if current expires for some reason", func() {
			koofr.InvalidateToken(client.GetToken())
			m, err := client.Mounts()
			Expect(err).NotTo(HaveOccurred())
			Expect(m).To(HaveLen(0))
		})

		It("should list devices", func() {
			d, err := client.Devices()
			Expect(err).NotTo(HaveOccurred())
			Expect(d).To(HaveLen(0))
		})

		It("should create device", func() {
			d, err := client.DevicesCreate("testDevice", "storagehub")
			Expect(err).NotTo(HaveOccurred())
			Expect(d).NotTo(BeNil())
		})

		It("should fetch device details", func() {
			d, err := client.DevicesDetails("testDeviceId")
			Expect(err).NotTo(HaveOccurred())
			Expect(d).NotTo(BeNil())
		})

	})

})
