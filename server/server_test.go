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

	t.Run("should return 422 with error body for invalid json request", func(t *testing.T) {
		invalidBodies := [...]string{"", "{"}
		for _, b := range invalidBodies {
			req, resp := makeCreateArticleRawRequestSuite(b)
			server.ServeHTTP(resp, req)
			assert422(t, resp)
		}
	})

	t.Run("should return 422 with error body for missing required fields", func(t *testing.T) {
		testCases := [...]struct {
			a        Article
			required []string
		}{
			{Article{}, []string{"Title"}},
		}
		for _, tc := range testCases {
			req, resp := makeCreateArticleRequestSuite(tc.a)
			server.ServeHTTP(resp, req)
			body := assert422(t, resp)
			assertRequiredFields(t, tc.a, tc.required, body)
		}
	})

	//TODO: create test to check response on store level creation error
}

//TODO: create test for unsupported routes

func setAuth(r *http.Request) {
	r.Header.Add("Authorization", "Bearer "+CreateToken(AuthData{"user1"}))
}

func makeGetArticleRequestSuite(slug string) (*http.Request, *httptest.ResponseRecorder) {
	req, _ := http.NewRequest(http.MethodGet, "/api/articles/"+slug, nil)
	return req, httptest.NewRecorder()
}

func makeCreateArticleRequestSuite(a Article) (*http.Request, *httptest.ResponseRecorder) {
	serializedArticle, _ := json.Marshal(SingleArticleHTTPWrap{a})
	req, _ := http.NewRequest(http.MethodPost, "/api/articles", bytes.NewBuffer(serializedArticle))
	setAuth(req)
	return req, httptest.NewRecorder()
}

func makeCreateArticleRawRequestSuite(body string) (*http.Request, *httptest.ResponseRecorder) {
	req, _ := http.NewRequest(http.MethodPost, "/api/articles", bytes.NewBuffer([]byte(body)))
	setAuth(req)
	return req, httptest.NewRecorder()
}

func assertStatus(t *testing.T, want, got int, message string) {
	t.Helper()
	if want != got {
		assert.FailNow(t, fmt.Sprintf("didn't get correct status. got %d instead of %d. %s", got, want, message))
	}
}

func assertJSONContentType(t *testing.T, resp *httptest.ResponseRecorder) {
	t.Helper()
	assert.Equal(t, "application/json", resp.Result().Header.Get("content-type"), "invalid content-type")
}

func assertJSONBody(t *testing.T, gotJSON string, compareTo interface{}, msg string) {
	t.Helper()
	builder := strings.Builder{}
	json.NewEncoder(&builder).Encode(compareTo)
	assert.Equal(t, builder.String(), gotJSON, msg)
}

func assertSussessJSONResponse(t *testing.T, resp *httptest.ResponseRecorder, bodyCompareTo interface{}) {
	t.Helper()
	assertStatus(t, http.StatusOK, resp.Code, "on valid request")
	assertJSONContentType(t, resp)
	assertJSONBody(t, resp.Body.String(), bodyCompareTo, "response body doesnt match desired struct")
	//TODO: think about object comparison instead of strings
	//TODO: compare desired json string from file system. it will show serialization errors (currently there are no json tags in structs)
}

func assert422(t *testing.T, resp *httptest.ResponseRecorder) UnprocessableEntityResponse {
	t.Helper()
	assertStatus(t, http.StatusUnprocessableEntity, resp.Code, fmt.Sprintf("on invalid request"))
	assertJSONContentType(t, resp)
	var body UnprocessableEntityResponse
	err := json.NewDecoder(resp.Body).Decode(&body)
	failOnNotEqual(t, err, nil, "expected to decode 422 response body without errors")
	failOnEqual(t, len(body.Errors.Body), 0, "expected to find elements in 422 response body block")
	return body
}

func assertRequiredFields(t *testing.T, requiredSource interface{}, requiredFields []string, response UnprocessableEntityResponse) {
	missing := []string{}
	for _, r := range requiredFields {
		r = strings.ToLower(r)
		var isFound bool
		for _, e := range response.Errors.Body {
			if strings.Contains(strings.ToLower(e), r) {
				isFound = true
				break
			}
		}
		if !isFound {
			missing = append(missing, r)
		}
	}
	if len(missing) > 0 {
		assert.Fail(t, fmt.Sprintf(
			"expected to see missing required fields in response body\nmissing: %q\nrequest data: %+v\nresponse data: %+v",
			missing, requiredSource, response))
	}
}
