package simple

import (
	"encoding/json"
	"os"

	"github.com/QuangTung97/weblib/auth/oauth"
	"github.com/QuangTung97/weblib/urls"
)

type GoogleClient struct {
	Web GoogleClientWeb `json:"web"`
}

type GoogleClientWeb struct {
	ClientID  string `json:"client_id"`
	ProjectID string `json:"project_id"`
	AuthURI   string `json:"auth_uri"`
	TokenURI  string `json:"token_uri"`

	CertURI string `json:"auth_provider_x509_cert_url"`

	ClientSecret string `json:"client_secret"`

	RedirectURIs []string `json:"redirect_uris"`
	JsOrigins    []string `json:"javascript_origins"`
}

func LoadGoogleClient(filePath string) GoogleClient {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer func() { _ = file.Close() }()

	dec := json.NewDecoder(file)
	dec.DisallowUnknownFields()

	var result GoogleClient
	if err := dec.Decode(&result); err != nil {
		panic(err)
	}
	return result
}

var LoginPath = urls.New[oauth.LoginParams]("/login")

var OauthCallbackPath = urls.New[oauth.CallbackParams]("/oauth/google/callback")

type HomeParams struct {
}

var HomePath = urls.New[HomeParams]("/")
