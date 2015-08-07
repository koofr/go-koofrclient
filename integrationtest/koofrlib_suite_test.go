package integrationtest_test

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/koofr/go-koofrclient"
	"github.com/koofr/go-koofrclient/auth"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	client         *koofrclient.KoofrClient
	apiBase        string
	email          string
	password       string
	defaultMountId string
)

func oauthObtainer(url string) string {
	defer GinkgoRecover()
	codeCh := make(chan string, 0)

	fmt.Printf("Please open %s and confirm access\n", url)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		codeCh <- code
	})

	go http.ListenAndServe("localhost:1337", nil)

	return <-codeCh
}

func TestKoofrclient(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Koofrclient Suite")
}

var _ = BeforeSuite(func() {
	apiBase = os.Getenv("KOOFR_APIBASE")
	Expect(apiBase).NotTo(BeEmpty(), "Please provide KOOFR_APIBASE env")

	email = os.Getenv("KOOFR_EMAIL")
	password = os.Getenv("KOOFR_PASSWORD")

	clientId := os.Getenv("KOOFR_CLIENT_ID")
	clientSecret := os.Getenv("KOOFR_CLIENT_SECRET")

	client = koofrclient.NewKoofrClient(apiBase, true)

	var ap auth.AuthProvider

	switch {
	case email != "" && password != "":
		ap = auth.NewTokenAuthProvider(email, password)
	case clientId != "" && clientSecret != "":
		ap = auth.NewOAuth2Provider(clientId, clientSecret, []string{"public", "private"}, "http://localhost:1337", oauthObtainer)
	default:
		Fail("Please provide (KOOFR_EMAIL,KOOFR_PASSWORD) or (KOOFR_CLIENT_ID, KOOFR_CLIENT_SECRET)")
	}

	err := client.AuthenticateWithProvider(ap)
	Expect(err).NotTo(HaveOccurred(), "Auth failed")

	mounts, err := client.Mounts()
	Expect(err).NotTo(HaveOccurred(), "Koofr listing mounts failed")

	Expect(mounts).NotTo(HaveLen(0), "Koofr mounts must not be empty")

	for _, m := range mounts {
		if m.IsPrimary {
			defaultMountId = m.Id
		}
	}

	Expect(defaultMountId).NotTo(BeEmpty(), "Koofr primary mount not found")

})
