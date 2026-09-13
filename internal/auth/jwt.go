package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	return hash, err
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	return match, err
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(tokenSecret))
	return ss, err
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	registeredClaims := jwt.RegisteredClaims{}
	keyFunc := func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	}
	token, err := jwt.ParseWithClaims(tokenString, &registeredClaims, keyFunc)
	if err != nil {
		return uuid.UUID{}, err
	}
	claims := token.Claims
	id, err := claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, err
	}

	uuidID, err := uuid.Parse(id)
	return uuidID, err
}

func GetBearerToken(headers http.Header) (string, error) {
	authorization := headers.Get("Authorization")
	if strings.HasPrefix(authorization, "Bearer ") == false {
		return "", fmt.Errorf("Authorization header has wrong format.")
	}
	return strings.TrimPrefix(authorization, "Bearer "), nil
}
