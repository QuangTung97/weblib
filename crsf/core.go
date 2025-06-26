package crsf

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"strconv"
	"strings"
)

type Error struct {
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

type Core struct {
	secretKey string
	randFunc  func() ([]byte, error)
}

func NewCore(
	secretKey string, // for HMAC
	randFunc func() ([]byte, error),
) *Core {
	return &Core{
		secretKey: secretKey,
		randFunc:  randFunc,
	}
}

func (g *Core) generateBeforeDigest(sessionID string) ([]byte, string, error) {
	randData, err := g.randFunc()
	if err != nil {
		return nil, "", err
	}

	randStr := hex.EncodeToString(randData)
	return g.generateMessageWithRandData(sessionID, randStr), randStr, nil
}

func (g *Core) generateMessageWithRandData(sessionID string, randStr string) []byte {
	var msg bytes.Buffer

	// append session id
	sessLen := strconv.FormatInt(int64(len(sessionID)), 10)
	msg.WriteString(sessLen)
	msg.WriteString("!")
	msg.WriteString(sessionID)
	msg.WriteString("!")

	// append random data
	randLen := strconv.FormatInt(int64(len(randStr)), 10)
	msg.WriteString(randLen)
	msg.WriteString("!")
	msg.WriteString(randStr)

	return msg.Bytes()
}

func (g *Core) Generate(sessionID string) (string, error) {
	msgBytes, randStr, err := g.generateBeforeDigest(sessionID)
	if err != nil {
		return "", err
	}

	hashBytes := g.computeHMAC(msgBytes)
	return base64.StdEncoding.EncodeToString(hashBytes) + "." + randStr, nil
}

func (g *Core) computeHMAC(msgBytes []byte) []byte {
	h := hmac.New(sha256.New, []byte(g.secretKey))
	h.Write(msgBytes)
	return h.Sum(nil)
}

func (g *Core) Validate(sessionID string, csrfToken string) error {
	parts := strings.Split(csrfToken, ".")
	if len(parts) != 2 {
		return &Error{Message: "invalid csrf token format"}
	}

	digestPart := parts[0]
	randStr := parts[1]

	inputDigest, err := base64.StdEncoding.DecodeString(digestPart)
	if err != nil {
		return err
	}

	expectedMsgBytes := g.generateMessageWithRandData(sessionID, randStr)
	expectedDigest := g.computeHMAC(expectedMsgBytes)

	if subtle.ConstantTimeCompare(inputDigest, expectedDigest) != 1 {
		return &Error{Message: "invalid csrf token"}
	}

	return nil
}
