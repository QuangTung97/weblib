package oauth

import (
	"crypto/rand"
	"time"

	"golang.org/x/oauth2"
)

func InitService(conf *oauth2.Config, successCallback SuccessCallback) Service {
	return NewService(
		conf, successCallback,
		time.Now,
		cryptRandFunc,
	)
}

func cryptRandFunc(n int) []byte {
	data := make([]byte, n)
	if _, err := rand.Read(data); err != nil {
		panic(err)
	}
	return data
}
