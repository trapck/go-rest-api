package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGETArticle(t *testing.T) {
	request, _ := http.NewRequest(http.MethodGet, "/api/articles/some-art", nil)
	response := httptest.NewRecorder()

	BlogServer(response, request)

	got := response.Body.String()
	want := "some art"

	assert.Equal(t, want, got, "should return correct article")

	request, _ = http.NewRequest(http.MethodGet, "/api/articles/some-other-art", nil)
	response = httptest.NewRecorder()

	BlogServer(response, request)

	got = response.Body.String()
	want = "some other art"

	assert.Equal(t, want, got, "should return correct article")
}
