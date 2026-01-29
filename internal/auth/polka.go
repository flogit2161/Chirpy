package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	headerAuth := headers.Get("Authorization")
	if headerAuth == "" {
		return "", fmt.Errorf("Header is empty")
	}

	if !strings.HasPrefix(headerAuth, "ApiKey ") {
		return "", fmt.Errorf("Token has a wrong prefix")
	}

	key := strings.TrimPrefix(headerAuth, "ApiKey ")

	cleanKey := strings.TrimSpace(key)

	return cleanKey, nil
}
