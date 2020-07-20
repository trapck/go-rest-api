package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type StubBlogStore struct {
	articles []Article
	users    []RequestUserData
}

func (s *StubBlogStore) GetArticle(slug string) (article Article, e error) {
	e = fmt.Errorf("Article with slug %q was not found", slug)
	for _, a := range s.articles {
		if a.Slug == slug {
			article = a
			if article.AuthorID.Valid {
				u, _ := s.GetUserByID(int(article.AuthorID.Int32))
				article.Author = u.ToProfile()
			}
			e = nil
			break
		}
	}
	return
}

func (s *StubBlogStore) CreateArticle(a SingleArticleHTTPWrap) (Article, error) {
	a.Article.Slug = CreateSlug(a.Title)
	s.articles = append(s.articles, a.Article)
	if a.AuthorID.Valid {
		u, _ := s.GetUserByID(int(a.AuthorID.Int32))
		a.Author = u.ToProfile()
	}
	return a.Article, nil
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

func (s *StubBlogStore) GetUserByID(id int) (user RequestUserData, e error) {
	e = fmt.Errorf("User with id %q was not found", id)
	for _, u := range s.users {
		if u.ID == id {
			user = u
			e = nil
			break
		}
	}
	return
}

func (s *StubBlogStore) UpdateUser(username string, data RequestUserData) (u RequestUserData, e error) {
	for i := range s.users {
		if s.users[i].UserName == username {
			s.users[i].UserName = data.UserName
			s.users[i].Email = data.Email
			s.users[i].Password = data.Password
			s.users[i].Bio = data.Bio
			s.users[i].Image = data.Image
			u = s.users[i]
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
		Article{0, "some-art", "some art", sql.NullInt32{}, Profile{}},
		Article{1, "some-other-art", "some other art", sql.NullInt32{}, Profile{}},
	}
	server := NewBlogServer(&StubBlogStore{testCases, nil})

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
	article := Article{0, "new-art", "new art", sql.NullInt32{}, Profile{}}
	user := RequestUserData{CommonUserData: CommonUserData{ID: 5, UserName: "denis"}}
	store := &StubBlogStore{users: []RequestUserData{user}}
	server := NewBlogServer(store)

	t.Run("should return created article", func(t *testing.T) {
		req, resp := makeCreateArticleRequestSuite(article)
		setAuth(req, AuthData{user.UserName})
		server.ServeHTTP(resp, req)
		var createdArticle SingleArticleHTTPWrap
		assertSussessJSONResponse(t, resp, &createdArticle)
		failOnEqual(t, createdArticle.Slug, "", "expected created article to have a slug")
		assert.Equal(t, article.Title, createdArticle.Title, "response article must have expected title")
		assert.Equal(t, user.UserName, createdArticle.Author.UserName, "response article must have expected author")
		_, err := store.GetArticle(createdArticle.Slug)
		failOnNotEqual(
			t,
			err,
			nil,
			fmt.Sprintf("expected to find article with slug %q in store. got error %v", createdArticle.Slug, err),
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
			setAuth(req, AuthData{user.UserName})
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

func TestGetCurrentUser(t *testing.T) {
	username := "user1"
	user := RequestUserData{CommonUserData: CommonUserData{UserName: username}}
	store := &StubBlogStore{nil, []RequestUserData{user}}
	server := NewBlogServer(store)

	t.Run("should return current user by auth token", func(t *testing.T) {
		req, resp := makeGetCurrentUserRequestSuite(username)
		server.ServeHTTP(resp, req)
		var currentUser ResponseUser
		assertSussessJSONResponse(t, resp, &currentUser)
		assert.Equal(t, username, currentUser.User.UserName, "exepected current user to have expected username")
	})

	t.Run("should return 404 for not existing user", func(t *testing.T) {
		req, resp := makeGetCurrentUserRequestSuite(username + "123")
		server.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})
}

func TestAuthentication(t *testing.T) {
	username := "user1"
	password := "123"
	user := RequestUserData{CommonUserData: CommonUserData{UserName: username}, Password: password}
	store := &StubBlogStore{nil, []RequestUserData{user}}
	server := NewBlogServer(store)

	t.Run("should authenticate user by auth body data", func(t *testing.T) {
		req, resp := makeAuthenticationRequestSuite(user)
		server.ServeHTTP(resp, req)
		var authenticatedUser ResponseUser
		assertSussessJSONResponse(t, resp, &authenticatedUser)
		assert.NotEmpty(t, authenticatedUser.User.Token, "exepected authenticated user to have an auth token")
	})

	t.Run("should return 404 for not existing user", func(t *testing.T) {
		fakeUser := user
		fakeUser.UserName = user.UserName + "123"
		req, resp := makeAuthenticationRequestSuite(fakeUser)
		server.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("should return 404 for existing user with incorrect password", func(t *testing.T) {
		fakeUser := user
		fakeUser.Password = user.Password + "123"
		req, resp := makeAuthenticationRequestSuite(fakeUser)
		server.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("should return 422 with error body for invalid json request", func(t *testing.T) {
		invalidBodies := [...]string{"", "{"}
		for _, b := range invalidBodies {
			req, resp := makeAuthenticationRawRequestSuite(b)
			server.ServeHTTP(resp, req)
			assert422(t, resp)
		}
	})

	t.Run("should return 422 with error body for missing required fields", func(t *testing.T) {
		testCases := [...]struct {
			u        RequestUserData
			required []string
		}{
			{RequestUserData{}, []string{"UserName", "Password"}},
			{RequestUserData{CommonUserData: CommonUserData{UserName: "denis"}}, []string{"Password"}},
		}
		for _, tc := range testCases {
			req, resp := makeRegistrationRequestSuite(tc.u)
			server.ServeHTTP(resp, req)
			body := assert422(t, resp)
			assertRequiredFields(t, tc.u, tc.required, body)
		}
	})
}

func TestPutUser(t *testing.T) {
	t.Run("should return updated user", func(t *testing.T) {
		authData := AuthData{"u"}
		u := RequestUserData{CommonUserData: CommonUserData{UserName: "u1", Bio: "b", Image: "i", Email: "e"}, Password: "p"}
		updateUser := UpdateUserData{UserName: &(u.UserName), Email: &(u.Email), Password: &(u.Password), Bio: &(u.Bio), Image: &(u.Image)}
		store := &StubBlogStore{nil, []RequestUserData{RequestUserData{CommonUserData: CommonUserData{UserName: authData.Login}}}}
		server := NewBlogServer(store)
		req, resp := makeUpdateUserRequestSuite(updateUser)
		setAuth(req, authData)
		server.ServeHTTP(resp, req)

		var updatedUser ResponseUser
		assertSussessJSONResponse(t, resp, &updatedUser)
		storeUser, err := store.GetUser(u.UserName)
		failOnNotEqual(t, err, nil, fmt.Sprintf("expected to find user with updated username %q in store. got error %v", u.UserName, err))
		assert.Equal(t, u.Email, storeUser.Email, "user email was not update correctly")
		assert.Equal(t, u.Bio, storeUser.Bio, "user bio was not update correctly")
		assert.Equal(t, u.Image, storeUser.Image, "user image was not update correctly")
		assert.Equal(t, u.Password, storeUser.Password, "user password was not update correctly")
		parsedAuthData, _ := ParseToken(updatedUser.User.Token)
		assert.Equal(t, u.UserName, parsedAuthData.Login, "should return auth token for new username")
	})

	t.Run("should not clear user fields that are not in json", func(t *testing.T) {
		authData := AuthData{"u"}
		primaryStoreUser := RequestUserData{CommonUserData: CommonUserData{UserName: authData.Login, Bio: "b", Image: "i", Email: "e"}, Password: "p"}
		store := &StubBlogStore{nil, []RequestUserData{primaryStoreUser}}
		server := NewBlogServer(store)
		req, resp := makeUpdateUserRequestSuite(UpdateUserData{})
		setAuth(req, authData)
		server.ServeHTTP(resp, req)

		var updatedUser ResponseUser
		assertSussessJSONResponse(t, resp, &updatedUser)
		resultStoreUser, err := store.GetUser(authData.Login)
		failOnNotEqual(t, err, nil, fmt.Sprintf("expected to find user with old username %q in store. got error %v", authData.Login, err))
		assert.Equal(t, primaryStoreUser.Email, resultStoreUser.Email, "expected user email to be not changed")
		assert.Equal(t, primaryStoreUser.Bio, resultStoreUser.Bio, "user bio to be not changed")
		assert.Equal(t, primaryStoreUser.Image, resultStoreUser.Image, "user image to be not changed")
		assert.Equal(t, primaryStoreUser.Password, resultStoreUser.Password, "user password to be not changed")
	})

	t.Run("should return 404 for not existing user", func(t *testing.T) {
		authData := AuthData{"u"}
		store := &StubBlogStore{nil, []RequestUserData{}}
		server := NewBlogServer(store)
		req, resp := makeUpdateUserRequestSuite(UpdateUserData{})
		setAuth(req, authData)
		server.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("should return 422 with error body for invalid json request", func(t *testing.T) {
		server := NewBlogServer(&StubBlogStore{})
		invalidBodies := [...]string{"", "{"}
		for _, b := range invalidBodies {
			req, resp := makeUpdateUserRawRequestSuite(b)
			setAuth(req, AuthData{"u"})
			server.ServeHTTP(resp, req)
			assert422(t, resp)
		}
	})
}

//endregion

//TODO: create test for unsupported routes, invalid route + method pairs

//TODO: add auth test for routes with auth

//region utils

func setAuth(r *http.Request, a AuthData) {
	r.Header.Add(HeaderKeyAuthorization, AuthHeader0Part+" "+CreateToken(a))
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

func makeCreateArticleRawRequestSuite(body string) (*http.Request, *httptest.ResponseRecorder) {
	req, _ := http.NewRequest(http.MethodPost, "/api/articles", bytes.NewBuffer([]byte(body)))
	setAuth(req, AuthData{"user1"})
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

func makeGetCurrentUserRequestSuite(username string) (*http.Request, *httptest.ResponseRecorder) {
	req, _ := http.NewRequest(http.MethodGet, "/api/user", nil)
	setAuth(req, AuthData{username})
	return req, httptest.NewRecorder()
}

func makeAuthenticationRequestSuite(u RequestUserData) (*http.Request, *httptest.ResponseRecorder) {
	serializedUser, _ := json.Marshal(RequestUser{u})
	req, _ := http.NewRequest(http.MethodPost, "/api/users/login", bytes.NewBuffer(serializedUser))
	return req, httptest.NewRecorder()
}

func makeAuthenticationRawRequestSuite(body string) (*http.Request, *httptest.ResponseRecorder) {
	req, _ := http.NewRequest(http.MethodPost, "/api/users/login", bytes.NewBuffer([]byte(body)))
	return req, httptest.NewRecorder()
}

func makeUpdateUserRequestSuite(u UpdateUserData) (*http.Request, *httptest.ResponseRecorder) {
	serializedUser, _ := json.Marshal(UpdateUserRequest{User: u})
	req, _ := http.NewRequest(http.MethodPut, "/api/user", bytes.NewBuffer(serializedUser))
	return req, httptest.NewRecorder()
}

func makeUpdateUserRawRequestSuite(body string) (*http.Request, *httptest.ResponseRecorder) {
	req, _ := http.NewRequest(http.MethodPut, "/api/user", bytes.NewBuffer([]byte(body)))
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
