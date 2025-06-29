package main

import (
	"log/slog"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/QuangTung97/weblib/auth/oauth"
	"github.com/QuangTung97/weblib/hx"
	"github.com/QuangTung97/weblib/router"
	"github.com/QuangTung97/weblib/urls"
)

const oauthLoginPath = "/oauth/login"

var oauthCallbackPath = urls.New[oauth.CallbackParams]("/oauth/callback")

type HomeParams struct {
}

var homePath = urls.New[HomeParams]("/")

func main() {
	rootRouter := router.NewRouter()

	oauthConf := &oauth2.Config{
		RedirectURL:  "http://localhost:8080/auth/google/callback",
		ClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
	oauthSvc := oauth.InitService(oauthConf, oauth.GoogleSuccessCallback())

	// setup url paths
	rootRouter.GetChi().Get(oauthLoginPath, oauthSvc.HandleLogin)
	router.HtmlGet(rootRouter, oauthCallbackPath, oauthSvc.HandleCallback)
	router.HtmlGet(rootRouter, homePath, func(ctx router.Context, params HomeParams) (hx.Elem, error) {
		return hx.Div(
			hx.Text("Hello World"),
		), nil
	})

	slog.Info("Start listen on port :8080")
	if err := http.ListenAndServe(":8080", rootRouter.GetChi()); err != nil {
		panic(err)
	}
}
