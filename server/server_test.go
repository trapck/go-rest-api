package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type StubBlogStore struct {
	data            []Article
	createdArticles int
}

func (s *StubBlogStore) GetArticle(slug string) (article Article, e error) {
	e = fmt.Errorf("Article with slug %s was not found", slug)
	for _, a := range s.data {
		if a.Slug == slug {
			article = a
			e = nil
			break
		}
	}
	return article, e
}

func (s *StubBlogStore) CreateArticle(a SingleArticleHTTPWrap) (article Article, e error) {
	return a.Article, nil
}

func TestGetArticle(t *testing.T) {
	testCases := []Article{
		Article{"some-art", "some art"},
		Article{"some-other-art", "some other art"},
	}
	server := &BlogServer{&StubBlogStore{testCases, 0}}

	t.Run("should return correct articles", func(t *testing.T) {
		for _, a := range testCases {
			req, resp := makeGetArticleRequestSuite(a.Slug)
			server.ServeHTTP(resp, req)
			assertStatus(t, http.StatusOK, resp.Code)
			builder := strings.Builder{}
			json.NewEncoder(&builder).Encode(SingleArticleHTTPWrap{a})
			assert.Equal(t, builder.String(), resp.Body.String(), "missing correct article")
		}
	})

	t.Run("returns 404 on missing articles", func(t *testing.T) {
		req, resp := makeGetArticleRequestSuite("not-existing-art")
		server.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

}

func TestCreateArticle(t *testing.T) {
	article := Article{"new-art", "new art"}
	server := &BlogServer{&StubBlogStore{}}

	t.Run("should return created article", func(t *testing.T) {
		req, resp := makeCreateArticleRequestSuite(article)
		server.ServeHTTP(resp, req)
		assertStatus(t, http.StatusOK, resp.Code)
		builder := strings.Builder{}
		json.NewEncoder(&builder).Encode(SingleArticleHTTPWrap{article})
		assert.Equal(t, builder.String(), resp.Body.String(), "created article doesnt match request")
	})
}

func makeGetArticleRequestSuite(slug string) (*http.Request, *httptest.ResponseRecorder) {
	req, _ := http.NewRequest(http.MethodGet, "/api/articles/"+slug, nil)
	return req, httptest.NewRecorder()
}

func makeCreateArticleRequestSuite(a Article) (*http.Request, *httptest.ResponseRecorder) {
	serializedArticle, _ := json.Marshal(SingleArticleHTTPWrap{a})
	req, _ := http.NewRequest(http.MethodPost, "/api/articles", bytes.NewBuffer(serializedArticle))
	return req, httptest.NewRecorder()
}

func assertStatus(t *testing.T, want, got int) {
	t.Helper()
	assert.Equal(t, want, got, fmt.Sprintf("did not get correct status, got %d, want %d", got, want))
}
