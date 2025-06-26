package crsf

import (
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCore_Generate_Then_Validate(t *testing.T) {
	t.Run("generate normal", func(t *testing.T) {
		g := NewCore("secret-01", func() ([]byte, error) {
			return []byte("rand-data-01"), nil
		})

		output01, err := g.Generate("session01")
		assert.Equal(t, nil, err)
		assert.Equal(t, 44+25, len(output01))
		assert.Equal(t, true, strings.HasSuffix(output01, "."+hex.EncodeToString([]byte("rand-data-01"))))

		output02, err := g.Generate("longer session value")
		assert.Equal(t, nil, err)
		assert.Equal(t, 44+25, len(output02))

		// different secret
		g = NewCore("secret-02-value", func() ([]byte, error) {
			return []byte("rand-data-01"), nil
		})
		output03, err := g.Generate("session01")
		assert.Equal(t, nil, err)
		assert.Equal(t, 44+25, len(output03))

		// different random value
		g = NewCore("secret-01", func() ([]byte, error) {
			return []byte("rand-data-02-another"), nil
		})
		output04, err := g.Generate("session01")
		assert.Equal(t, nil, err)
		assert.Equal(t, 44+25+16, len(output04))

		// check all distinct
		set := map[string]struct{}{
			output01: {},
			output02: {},
			output03: {},
			output04: {},
		}
		assert.Equal(t, 4, len(set))
	})

	t.Run("check msg before digest", func(t *testing.T) {
		g := NewCore("secret-01", func() ([]byte, error) {
			return []byte("rand-data-01"), nil
		})

		msgBytes, randStr, err := g.generateBeforeDigest("session01")
		assert.Equal(t, nil, err)
		assert.Equal(t, hex.EncodeToString([]byte("rand-data-01")), randStr)
		assert.Equal(t,
			"9!session01!24!"+randStr,
			string(msgBytes),
		)
	})

	t.Run("generate with random error", func(t *testing.T) {
		g := NewCore("secret-01", func() ([]byte, error) {
			return nil, errors.New("random error")
		})

		output01, err := g.Generate("session01")
		assert.Equal(t, errors.New("random error"), err)
		assert.Equal(t, "", output01)
	})

	t.Run("generate then validate", func(t *testing.T) {
		g := NewCore("secret-01", func() ([]byte, error) {
			return []byte("rand-data-01"), nil
		})

		output01, err := g.Generate("session01")
		assert.Equal(t, nil, err)

		// ok
		err = g.Validate("session01", output01)
		assert.Equal(t, nil, err)

		// different session
		err = g.Validate("session02", output01)
		assert.Equal(t, &Error{Message: "invalid csrf token"}, err)

		// different secret key
		g = NewCore("secret-02-new-value", nil)
		err = g.Validate("session01", output01)
		assert.Equal(t, &Error{Message: "invalid csrf token"}, err)
		assert.Equal(t, "invalid csrf token", err.Error())
	})

	t.Run("validate invalid format, missing dot", func(t *testing.T) {
		g := NewCore("secret-01", nil)
		err := g.Validate("session01", "invalid-format-csrf-token")
		assert.Equal(t, &Error{Message: "invalid csrf token format"}, err)
	})

	t.Run("validate invalid format, not a base64 string", func(t *testing.T) {
		g := NewCore("secret-01", nil)
		err := g.Validate("session01", "invalid01.random02")
		assert.Error(t, err)
		assert.Equal(t, "illegal base64 data at input byte 8", err.Error())
	})
}
