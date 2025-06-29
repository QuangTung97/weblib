package oauth

import (
	"fmt"
	"io"
	"net/http"

	"github.com/QuangTung97/weblib/router"
)

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

func GoogleSuccessCallback() SuccessCallback {
	return func(ctx router.Context, accessToken string) error {
		resp, err := http.Get(oauthGoogleUrlAPI + accessToken)
		if err != nil {
			return fmt.Errorf("failed getting user info: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed reading user info: %w", err)
		}

		fmt.Println(string(data))
		return nil
	}
}
