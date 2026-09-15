package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	authorization := headers.Get("Authorization")
	if strings.HasPrefix(authorization, "ApiKey ") == false {
		return "", fmt.Errorf("Authorization header has wrong format.")
	}
	return strings.TrimPrefix(authorization, "ApiKey "), nil
}
