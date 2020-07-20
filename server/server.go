package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// BlogStore stores blog data
type BlogStore interface {
	GetArticle(search string) (Article, error)
	CreateArticle(a SingleArticleHTTPWrap) (Article, error)
	GetUser(username string) (RequestUserData, error)
	UpdateUser(username string, data RequestUserData) (RequestUserData, error)
	Registration(user RequestUserData) (RequestUserData, error)
}

// BlogServer handles bolg api requests
type BlogServer struct {
	Store BlogStore
	http.Handler
}

func (s *BlogServer) serveGetArticle(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/api/articles/")
	article, err := s.Store.GetArticle(slug)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	} else {
		writeJSONResponse(w, SingleArticleHTTPWrap{article})
	}
}

func (s *BlogServer) serveCreateArticle(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	if reqData, err := parseCreateArticleBody(body); err != nil {
		write422Response(w, err)
	} else {
		t, _ := TokenFromAuthHeader(r)
		authData, _ := ParseToken(t)
		u, e := s.Store.GetUser(authData.Login)
		if e == nil {
			reqData.AuthorID = sql.NullInt32{Int32: int32(u.ID), Valid: true}
		}
		createdArticle, _ := s.Store.CreateArticle(reqData)
		writeJSONResponse(w, createdArticle)
	}
}

func (s *BlogServer) serveRegistration(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	if user, err := parseRegistrationBody(body); err != nil {
		write422Response(w, err)
	} else {
		registeredUser, _ := s.Store.Registration(user.User)
		commonUserData := registeredUser.ToCommonUserData()
		responseUser := ResponseUser{
			User: ResponseUserData{
				CommonUserData: commonUserData,
				Token:          CreateToken(AuthData{Login: commonUserData.UserName}),
			},
		}
		writeJSONResponse(w, responseUser)
	}
}

func (s *BlogServer) serveUser(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.serveGetCurrentUser(w, r)
	case http.MethodPut:
		s.serveUpdateUser(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *BlogServer) serveGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	t, _ := TokenFromAuthHeader(r)
	authData, _ := ParseToken(t)
	u, e := s.Store.GetUser(authData.Login)
	if e != nil {
		w.WriteHeader(http.StatusNotFound)
	} else {
		writeJSONResponse(w, ResponseUser{
			User: ResponseUserData{
				CommonUserData: u.ToCommonUserData(),
				Token:          t,
			},
		})
	}
}

func (s *BlogServer) serveUpdateUser(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	t, _ := TokenFromAuthHeader(r)
	authData, _ := ParseToken(t)
	if requestUser, err := parseUpdateUserBody(body); err != nil {
		write422Response(w, err)
	} else {
		foundUser, e := s.Store.GetUser(authData.Login)
		if e != nil {
			w.WriteHeader(http.StatusNotFound)
		} else {
			if requestUser.User.UserName != nil {
				foundUser.UserName = *requestUser.User.UserName
			}
			if requestUser.User.Password != nil {
				foundUser.Password = *requestUser.User.Password
			}
			if requestUser.User.Email != nil {
				foundUser.Email = *requestUser.User.Email
			}
			if requestUser.User.Bio != nil {
				foundUser.Bio = *requestUser.User.Bio
			}
			if requestUser.User.Image != nil {
				foundUser.Image = *requestUser.User.Image
			}
			u, e := s.Store.UpdateUser(authData.Login, foundUser)
			if e != nil {
				write500Response(w, e)
			} else {
				writeJSONResponse(w, ResponseUser{
					User: ResponseUserData{
						CommonUserData: u.ToCommonUserData(),
						Token:          CreateToken(AuthData{u.UserName}),
					},
				})
			}
		}
	}
}

func (s *BlogServer) serveAuthentication(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	if user, err := parseAuthenticationBody(body); err != nil {
		write422Response(w, err)
	} else {
		authenticatedUser, err := s.Store.GetUser(user.User.UserName)
		if err != nil || authenticatedUser.Password != user.User.Password {
			w.WriteHeader(http.StatusNotFound)
		} else {
			commonUserData := authenticatedUser.ToCommonUserData()
			responseUser := ResponseUser{
				User: ResponseUserData{
					CommonUserData: commonUserData,
					Token:          CreateToken(AuthData{Login: commonUserData.UserName}),
				},
			}
			writeJSONResponse(w, responseUser)
		}
	}
}

func (s *BlogServer) getRoutes() map[string]func(http.ResponseWriter, *http.Request) {
	return map[string]func(http.ResponseWriter, *http.Request){
		"/api/articles/":   s.serveGetArticle,
		"/api/articles":    s.serveCreateArticle,
		"/api/user":        s.serveUser,
		"/api/users/login": s.serveAuthentication,
		"/api/users":       s.serveRegistration,
	}
}

// NewBlogServer initializes new instance of the blog server
func NewBlogServer(s BlogStore) *BlogServer {
	server := BlogServer{Store: s}
	router := http.NewServeMux()
	for r, h := range server.getRoutes() {
		var handler http.Handler = http.HandlerFunc(h)
		if needAuth(r) {
			handler = ApplyAuth(handler)
		}
		router.Handle(r, handler)
	}
	server.Handler = router
	return &server
}

func needAuth(route string) bool {
	return route == "/api/articles" || route == "/api/user"
}

func parseRegistrationBody(b []byte) (data RequestUser, e error) {
	errors := []string{}
	decodeError := json.NewDecoder(bytes.NewBuffer(b)).Decode(&data)

	if decodeError != nil {
		errors = append(errors, MsgInvalidBody)
	} else { //TODO: use reflect
		missing := []string{}
		if data.User.UserName == "" {
			missing = append(missing, "UserName")
		}
		if data.User.Email == "" {
			missing = append(missing, "Email")
		}
		if data.User.Password == "" {
			missing = append(missing, "Password")
		}
		if len(missing) > 0 {
			errors = append(errors, fmt.Sprintf("Missing required fields: %q", strings.Join(missing, ",")))
		}
	}

	if len(errors) > 0 {
		e = &UnprocessableEntityResponse{Errors: UnprocessableEntityError{Body: errors}}
	}
	return data, e
}

func parseUpdateUserBody(b []byte) (data UpdateUserRequest, e error) {
	decodeError := json.NewDecoder(bytes.NewBuffer(b)).Decode(&data)
	if decodeError != nil {
		e = &UnprocessableEntityResponse{Errors: UnprocessableEntityError{Body: []string{MsgInvalidBody}}}
	}
	return data, e
}

func parseAuthenticationBody(b []byte) (data RequestUser, e error) {
	errors := []string{}
	decodeError := json.NewDecoder(bytes.NewBuffer(b)).Decode(&data)

	if decodeError != nil {
		errors = append(errors, MsgInvalidBody)
	} else { //TODO: use reflect
		missing := []string{}
		if data.User.UserName == "" {
			missing = append(missing, "UserName")
		}
		if data.User.Password == "" {
			missing = append(missing, "Password")
		}
		if len(missing) > 0 {
			errors = append(errors, fmt.Sprintf("Missing required fields: %q", strings.Join(missing, ",")))
		}
	}

	if len(errors) > 0 {
		e = &UnprocessableEntityResponse{Errors: UnprocessableEntityError{Body: errors}}
	}
	return data, e
}

func parseCreateArticleBody(b []byte) (data SingleArticleHTTPWrap, e error) {
	errors := []string{}
	decodeError := json.NewDecoder(bytes.NewBuffer(b)).Decode(&data)

	if decodeError != nil {
		errors = append(errors, MsgInvalidBody)
	} else if data.Article.Title == "" { //TODO: use reflect
		errors = append(errors, fmt.Sprintf("Missing required fields: %q", []string{"Title"}))
	}

	if len(errors) > 0 {
		e = &UnprocessableEntityResponse{Errors: UnprocessableEntityError{Body: errors}}
	}
	return data, e
}

func writeJSONContentType(w http.ResponseWriter) {
	w.Header().Set(HeaderKeyContentType, HeaderValueJSONContactType)
}

func writeJSONResponse(w http.ResponseWriter, v interface{}) {
	writeJSONContentType(w)
	json.NewEncoder(w).Encode(v)
}

func write422Response(w http.ResponseWriter, e error) {
	writeJSONContentType(w)
	w.WriteHeader(http.StatusUnprocessableEntity)
	w.Write([]byte(e.Error()))
}

func write500Response(w http.ResponseWriter, e error) {
	writeJSONContentType(w)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(e.Error()))
}
