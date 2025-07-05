package oauth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/QuangTung97/weblib/router"
)

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

type GoogleAccount struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Picture       string `json:"picture"`
}

func GoogleCallback(
	onHandler func(ctx router.Context, account GoogleAccount) error,
) SuccessCallback {
	return func(ctx router.Context, accessToken string) error {
		resp, err := http.Get(oauthGoogleUrlAPI + accessToken)
		if err != nil {
			return fmt.Errorf("failed getting user info: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		var account GoogleAccount
		if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
			return fmt.Errorf("failed decoding user info: %w", err)
		}

		return onHandler(ctx, account)
	}
}
