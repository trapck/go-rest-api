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
	articles        []Article
	users           []RequestUserData
	createdArticles int
}

func (s *StubBlogStore) GetArticle(slug string) (article Article, e error) {
	e = fmt.Errorf("Article with slug %q was not found", slug)
	for _, a := range s.articles {
		if a.Slug == slug {
			article = a
			e = nil
			break
		}
	}
	return
}

func (s *StubBlogStore) CreateArticle(a SingleArticleHTTPWrap) (Article, error) {
	newArticle := Article{Slug: CreateSlug(a.Title), Title: a.Title}
	s.articles = append(s.articles, newArticle)
	return newArticle, nil
}

func (s *StubBlogStore) GetUser(username string) (user RequestUserData, e error) {
	e = fmt.Errorf("User with username %q was not found", username)
	for _, u := range s.users {
		if u.UserName == username {
			user = u
			e = nil
			break
		}
	}
	return
}

func (s *StubBlogStore) Registration(user RequestUserData) (RequestUserData, error) {
	s.users = append(s.users, user)
	return user, nil
}

//region article

func TestGetArticle(t *testing.T) {
	testCases := []Article{
		Article{"some-art", "some art"},
		Article{"some-other-art", "some other art"},
	}
	server := NewBlogServer(&StubBlogStore{testCases, nil, 0})

	t.Run("should return correct article by search value", func(t *testing.T) {
		for _, a := range testCases {
			req, resp := makeGetArticleRequestSuite(a.Slug)
			server.ServeHTTP(resp, req)
			assertSussessJSONResponseExact(t, resp, SingleArticleHTTPWrap{a})
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
	store := &StubBlogStore{}
	server := NewBlogServer(store)

	t.Run("should return created article", func(t *testing.T) {
		req, resp := makeCreateArticleRequestSuite(article)
		server.ServeHTTP(resp, req)
		var createdArticle SingleArticleHTTPWrap
		assertSussessJSONResponse(t, resp, &createdArticle)
		failOnEqual(t, createdArticle.Slug, "", "expected created article to have a slug")
		storeArticle, err := store.GetArticle(createdArticle.Slug)
		failOnNotEqual(
			t,
			err,
			nil,
			fmt.Sprintf("expected to find article with slug %q in store. got error %v", createdArticle.Slug, err),
		)
		assert.Equal(t,
			article.Title,
			storeArticle.Title,
			fmt.Sprintf("found created article with slug %q should have expected title", createdArticle.Slug),
		)
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

//endregion

//region user

func TestRegistration(t *testing.T) {
	user := RequestUserData{CommonUserData: CommonUserData{UserName: "denis", Email: "denis@gmail.com"}, Password: "123"}
	store := &StubBlogStore{}
	server := NewBlogServer(store)

	t.Run("should return registered user", func(t *testing.T) {
		req, resp := makeRegistrationRequestSuite(user)
		server.ServeHTTP(resp, req)
		var registeredUser ResponseUser
		assertSussessJSONResponse(t, resp, &registeredUser)
		failOnEqual(t, "", registeredUser.User.Token, "expected registered user to have an auth token")
		storeUser, err := store.GetUser(user.UserName)
		failOnNotEqual(
			t,
			err,
			nil,
			fmt.Sprintf("expected to find user with username %q in store. got error %v", user.UserName, err),
		)
		assert.Equal(t,
			user.Email,
			storeUser.Email,
			fmt.Sprintf("found created user with username %q should have expected email", registeredUser.User.UserName),
		)
	})

	t.Run("should return 422 with error body for invalid json request", func(t *testing.T) {
		invalidBodies := [...]string{"", "{"}
		for _, b := range invalidBodies {
			req, resp := makeRegistrationRawRequestSuite(b)
			server.ServeHTTP(resp, req)
			assert422(t, resp)
		}
	})

	t.Run("should return 422 with error body for missing required fields", func(t *testing.T) {
		testCases := [...]struct {
			u        RequestUserData
			required []string
		}{
			{RequestUserData{}, []string{"UserName", "Email", "Password"}},
			{RequestUserData{CommonUserData: CommonUserData{UserName: "denis"}}, []string{"Email", "Password"}},
			{RequestUserData{CommonUserData: CommonUserData{UserName: "denis", Email: "denis@gmail.com"}}, []string{"Password"}},
		}
		for _, tc := range testCases {
			req, resp := makeRegistrationRequestSuite(tc.u)
			server.ServeHTTP(resp, req)
			body := assert422(t, resp)
			assertRequiredFields(t, tc.u, tc.required, body)
		}
	})

	//TODO: test invalid email
	//TODO: create test to check response on store level creation error
}

//endregion

//TODO: create test for unsupported routes

//region utils

func setAuth(r *http.Request) {
	r.Header.Add(HeaderKeyAuthorization, AuthHeader0Part+" "+CreateToken(AuthData{"user1"}))
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

func makeRegistrationRequestSuite(u RequestUserData) (*http.Request, *httptest.ResponseRecorder) {
	serializedUser, _ := json.Marshal(RequestUser{u})
	req, _ := http.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(serializedUser))
	return req, httptest.NewRecorder()
}

func makeRegistrationRawRequestSuite(body string) (*http.Request, *httptest.ResponseRecorder) {
	req, _ := http.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer([]byte(body)))
	return req, httptest.NewRecorder()
}

func assertStatus(t *testing.T, want, got int, message string) {
	t.Helper()
	failOnNotEqual(t, want, got, fmt.Sprintf("didn't get correct status. got %d instead of %d. %s", got, want, message))
}

func assertJSONContentType(t *testing.T, resp *httptest.ResponseRecorder) {
	t.Helper()
	got, want := resp.Result().Header.Get(HeaderKeyContentType), HeaderValueJSONContactType
	failOnNotEqual(t, want, got, fmt.Sprintf("invalid content-type. got %q instead of %q", got, want))
}

func assertSuccessJSONResponseHeaders(t *testing.T, resp *httptest.ResponseRecorder) {
	t.Helper()
	assertStatus(t, http.StatusOK, resp.Code, "on valid request")
	assertJSONContentType(t, resp)
}

func assertJSONBody(t *testing.T, gotJSON string, compareTo interface{}, msg string) {
	t.Helper()
	builder := strings.Builder{}
	json.NewEncoder(&builder).Encode(compareTo)
	assert.Equal(t, builder.String(), gotJSON, msg)
}

func assertSussessJSONResponseExact(t *testing.T, resp *httptest.ResponseRecorder, bodyCompareTo interface{}) {
	t.Helper()
	assertSuccessJSONResponseHeaders(t, resp)
	assertJSONBody(t, resp.Body.String(), bodyCompareTo, "response body doesnt match desired struct")
	//TODO: think about object comparison instead of strings
	//TODO: compare desired json string from file system. it will show serialization errors (currently there are no json tags in structs)
}

func assertSussessJSONResponse(t *testing.T, resp *httptest.ResponseRecorder, decodeTo interface{}) {
	t.Helper()
	assertSuccessJSONResponseHeaders(t, resp)
	err := json.NewDecoder(resp.Body).Decode(decodeTo)
	failOnNotEqual(t, err, nil, fmt.Sprintf("unable to decode %q to type %T. Got error %q", resp.Body.String(), decodeTo, err))
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
		var isFound bool
		for _, e := range response.Errors.Body {
			if strings.Contains(strings.ToLower(e), strings.ToLower(r)) {
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

//endregion
