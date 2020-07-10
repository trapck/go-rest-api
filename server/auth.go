package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
)

// AuthData is data model to use for auth token
type AuthData struct {
	Login string
}

// ApplyAuth applies midleware to check authorization by jwt
func ApplyAuth(next http.Handler) http.Handler {
	return jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte(authSecretKey), nil
		},
		Extractor:     TokenFromAuthHeader,
		SigningMethod: jwt.SigningMethodHS256,
	}).Handler(next)
}

// CreateToken generates auth token
func CreateToken(d AuthData) string {
	t, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": d,
		"exp":  time.Now().Add(30 * time.Minute).Unix(),
		"iat":  time.Now().Unix(),
	}).SignedString([]byte(authSecretKey))
	return t
}

// TokenFromAuthHeader extracts token from auth header
func TokenFromAuthHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", nil
	}

	authHeaderParts := strings.Fields(authHeader)
	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != strings.ToLower(AuthHeader0Part) {
		return "", fmt.Errorf("Authorization header format must be %q format", AuthHeader0Part)
	}

	return authHeaderParts[1], nil
}
