package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/QuangTung97/weblib/auth/oauth"
	"github.com/QuangTung97/weblib/csrf"
	"github.com/QuangTung97/weblib/examples/googlelogin/simple"
	"github.com/QuangTung97/weblib/hx"
	"github.com/QuangTung97/weblib/router"
)

func main() {
	rootRouter := router.NewRouter()
	rootRouter = rootRouter.WithMiddlewares(
		csrf.InitMiddleware("test-hmac-key01"),
	)

	clientFile := simple.LoadGoogleClient("data/google_client.json")

	oauthConf := &oauth2.Config{
		RedirectURL:  clientFile.Web.RedirectURIs[0],
		ClientID:     clientFile.Web.ClientID,
		ClientSecret: clientFile.Web.ClientSecret,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}

	callbackHandler := func(ctx router.Context, account oauth.GoogleAccount) error {
		fmt.Printf("HELLO: %+v\n", account)
		return nil
	}

	oauthSvc := oauth.InitService(
		oauthConf,
		oauth.GoogleCallback(callbackHandler),
	)

	// setup url paths
	router.HtmlGet(rootRouter, simple.LoginPath, oauthSvc.HandleLogin)
	router.HtmlGet(rootRouter, simple.OauthCallbackPath, oauthSvc.HandleCallback)

	// setup home page
	homeHandler := func(ctx router.Context, params simple.HomeParams) (hx.Elem, error) {
		loginPath := simple.LoginPath.Eval(oauth.LoginParams{
			Redirect: "/",
		})

		return hx.Div(
			hx.Text("Hello World"),
			hx.Br(),
			hx.A(
				hx.Text("Login with Google"),
				hx.Href(loginPath),
			),
		), nil
	}
	router.HtmlGet(rootRouter, simple.HomePath, homeHandler)

	slog.Info("Start listen on port :8080")
	if err := http.ListenAndServe(":8080", rootRouter.GetChi()); err != nil {
		panic(err)
	}
}
