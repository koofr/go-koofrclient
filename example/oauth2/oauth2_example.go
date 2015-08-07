package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/codegangsta/cli"
	"github.com/koofr/go-koofrclient"
	"github.com/koofr/go-koofrclient/auth"
)

func main() {
	fmt.Println("Oauth2 example")

	app := cli.NewApp()
	app.Name = "Koofr OAuth2"
	app.Usage = "example program how to use OAuth2 with Koofr"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "url",
			Value: "https://app.koofr.net",
			Usage: "URL of Koofr deploy to test against",
		},
		cli.StringFlag{
			Name:  "id",
			Value: "EJGOEWW46Q7CEIY7H7VTKYXUBU7DQ72P",
			Usage: "clientId of OAuth2 app",
		},
		cli.StringFlag{
			Name:  "secret",
			Value: "VVTHJKR5YH6DJ7FIYCM56FHJ54HVBUITT2WDVYUB652T2TR2H74WHZIJBB7OEFUX",
			Usage: "clientSecret of OAuth2 app",
		},
		cli.StringSliceFlag{
			Name:  "scopes",
			Value: &cli.StringSlice{"public"},
			Usage: "Required scopes",
		},
	}
	app.Action = do

	app.Run(os.Args)

}
func do(ctx *cli.Context) {

	c := koofrclient.NewKoofrClient(ctx.String("url"), true)

	ap := auth.NewOAuth2Provider(ctx.String("id"), ctx.String("secret"), []string{"public"}, "http://localhost:1337", obtainer)

	err := c.AuthenticateWithProvider(ap)
	if err != nil {
		panic(err)
	}

	mounts, err := c.Mounts()
	if err != nil {
		panic(err)
	}

	fmt.Println("List of your mounts")
	fmt.Println("=====================================")
	for _, m := range mounts {
		fmt.Printf("%s: %s\n", m.Id, m.Name)
	}
}

func obtainer(url string) string {

	codeCh := make(chan string, 0)

	fmt.Printf("Please open %s and confirm access\n", url)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		codeCh <- code
	})

	go http.ListenAndServe("localhost:1337", nil)

	return <-codeCh
}
