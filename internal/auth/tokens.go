package auth

import (
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
