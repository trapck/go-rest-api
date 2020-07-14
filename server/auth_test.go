package server

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateToken(t *testing.T) {
	u1 := AuthData{"user1"}
	u2 := AuthData{"user2"}
	t.Run("must generate unique token per user", func(t *testing.T) {
		assert.NotEqual(t, CreateToken(u1), CreateToken(u2), fmt.Sprintf("have two equal tokens for %+v and %+v", u1, u2))
	})
}

func TestParseToken(t *testing.T) {
	u := AuthData{"user1"}
	token := CreateToken(u)
	parsedU, err := ParseToken(token)
	failOnNotEqual(t, err, nil, fmt.Sprintf("expected to parse token %q without errors but gon %q", token, err))
	assert.Equal(t, u, parsedU, "parsed token must be equal to expected struct")
}

func TestTokenFromAuthHeader(t *testing.T) {
	validHeader := AuthHeader0Part + " " + "jwt"
	r, _ := http.NewRequest(http.MethodGet, "", nil)
	r.Header.Add(HeaderKeyAuthorization, validHeader)
	h, e := TokenFromAuthHeader(r)
	assert.NotEqual(t, "", h, "expected to get token from valid header")

	invalidHeader := AuthHeader0Part + "123 " + "jwt"
	r, _ = http.NewRequest(http.MethodGet, "", nil)
	r.Header.Add(HeaderKeyAuthorization, invalidHeader)
	h, e = TokenFromAuthHeader(r)
	assert.Error(t, e, "expected to get error for invalid header")

	r, _ = http.NewRequest(http.MethodGet, "", nil)
	h, e = TokenFromAuthHeader(r)
	assert.NoError(t, e, "expected to get no error for empty auth header")
	assert.Empty(t, h, "expected to get empty token from empty auth header")
}
