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
	return
}

func (s *StubBlogStore) CreateArticle(a SingleArticleHTTPWrap) (article Article, e error) {
	return a.Article, nil
}

func TestGetArticle(t *testing.T) {
	testCases := []Article{
		Article{"some-art", "some art"},
		Article{"some-other-art", "some other art"},
	}
	server := NewBlogServer(&StubBlogStore{testCases, 0})

	t.Run("should return correct article by search value", func(t *testing.T) {
		for _, a := range testCases {
			req, resp := makeGetArticleRequestSuite(a.Slug)
			server.ServeHTTP(resp, req)
			assertSussessJSONResponse(t, resp, SingleArticleHTTPWrap{a})
		}
	})

	t.Run("should return 404 on missing article by search value", func(t *testing.T) {
		req, resp := makeGetArticleRequestSuite("not-existing-art")
		server.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

}

func TestCreateArticle(t *testing.T) {
	article := Article{"new-art", "new art"}
	server := NewBlogServer(&StubBlogStore{})

	t.Run("should return created article", func(t *testing.T) {
		req, resp := makeCreateArticleRequestSuite(article)
		server.ServeHTTP(resp, req)
		assertSussessJSONResponse(t, resp, SingleArticleHTTPWrap{article})
	})

	//TODO: create test with invalid json in request body
	//TODO: create test to check response on store creation error
}

//TODO: create test for unsupported routes

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
	assert.Equal(t, want, got, "did not get correct status")
}

func assertJSONContentType(t *testing.T, got string) {
	t.Helper()
	assert.Equal(t, "application/json", got, "invalid content-type")
}

func assertJSONBody(t *testing.T, gotJSON string, compareTo interface{}, msg string) {
	t.Helper()
	builder := strings.Builder{}
	json.NewEncoder(&builder).Encode(compareTo)
	assert.Equal(t, builder.String(), gotJSON, msg)
}

func assertSussessJSONResponse(t *testing.T, resp *httptest.ResponseRecorder, bodyCompareTo interface{}) {
	t.Helper()
	assertStatus(t, http.StatusOK, resp.Code)
	assertJSONContentType(t, resp.Result().Header.Get("content-type"))
	assertJSONBody(t, resp.Body.String(), bodyCompareTo, "response body doesnt match desired struct")
	//TODO: think about object comparison instead of strings
	//TODO: think about test that will compare desired json string from file system with response body
}
