package csrf

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
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
	randFunc  func(n int) []byte
}

func InitCore(secretKey string) *Core {
	return NewCore(secretKey, realRand)
}

func realRand(n int) []byte {
	data := make([]byte, n)
	if _, err := rand.Read(data[:]); err != nil {
		panic(err)
	}
	return data
}

func NewCore(
	secretKey string, // for HMAC
	randFunc func(n int) []byte,
) *Core {
	return &Core{
		secretKey: secretKey,
		randFunc:  randFunc,
	}
}

func (g *Core) generateBeforeDigest(sessionID string) ([]byte, string) {
	randData := g.randFunc(20)
	randStr := hex.EncodeToString(randData)
	return g.generateMessageWithRandData(sessionID, randStr), randStr
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

func (g *Core) Generate(sessionID string) string {
	msgBytes, randStr := g.generateBeforeDigest(sessionID)
	hashBytes := g.computeHMAC(msgBytes)
	return hex.EncodeToString(hashBytes) + "." + randStr
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

	inputDigest, err := hex.DecodeString(digestPart)
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
