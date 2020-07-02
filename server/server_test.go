package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type StubBlogStore struct {
	data map[string]string
}

func (s *StubBlogStore) GetArticle(slug string) string {
	return s.data[slug]
}

func TestGETArticle(t *testing.T) {
	testCases := map[string]string{
		"some-art":       "some art",
		"some-other-art": "some other art",
	}
	server := &BlogServer{&StubBlogStore{testCases}}
	for key, value := range testCases {
		req, resp := createGETArticleRequestSuite(key)

		server.ServeHTTP(resp, req)
		got := resp.Body.String()
		want := value
		assert.Equal(t, want, got, "should return correct article")
	}
}

func createGETArticleRequestSuite(slug string) (*http.Request, *httptest.ResponseRecorder) {
	req, _ := http.NewRequest(http.MethodGet, "/api/articles/"+slug, nil)
	return req, httptest.NewRecorder()
}
