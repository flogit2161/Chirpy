package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
)

func GetBearerToken(headers http.Header) (string, error) {
	headerAuth := headers.Get("Authorization")
	if headerAuth == "" {
		return "", fmt.Errorf("Header is empty")
	}

	if !strings.HasPrefix(headerAuth, "Bearer ") {
		return "", fmt.Errorf("Token has a wrong prefix")
	}

	token := strings.TrimPrefix(headerAuth, "Bearer ")

	cleanToken := strings.TrimSpace(token)

	return cleanToken, nil

}

func MakeRefreshToken() (string, error) {
	byterand := make([]byte, 32)
	_, err := rand.Read(byterand)
	if err != nil {
		return "", fmt.Errorf("Error creating the random encoded string")
	}

	encodedString := hex.EncodeToString(byterand)
	return encodedString, nil
}
