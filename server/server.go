package server

import (
	"bytes"
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
	var reqData SingleArticleHTTPWrap
	body, _ := ioutil.ReadAll(r.Body)

	if err := validateCreateArticleBody(body); err != nil {
		writeJSONContentType(w)
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(UnprocessableEntityResponse{Errors: UnprocessableEntityError{[]string{MsgInvalidBody}}})
		return
	}

	json.NewDecoder(bytes.NewBuffer(body)).Decode(&reqData)
	createdArticle, _ := s.Store.CreateArticle(reqData)
	writeJSONResponse(w, createdArticle)
}

func (s *BlogServer) getRoutes() map[string]func(http.ResponseWriter, *http.Request) {
	return map[string]func(http.ResponseWriter, *http.Request){
		"/api/articles/": s.serveGetArticle,
		"/api/articles":  s.serveCreateArticle,
	}
}

// NewBlogServer initializes new instance of the blog server
func NewBlogServer(s BlogStore) *BlogServer {
	server := BlogServer{Store: s}
	router := http.NewServeMux()
	for r, h := range server.getRoutes() {
		router.HandleFunc(r, h)
	}
	server.Handler = router
	return &server
}

func validateCreateArticleBody(b []byte) (e error) {
	if !json.Valid(b) {
		e = fmt.Errorf("invalid json body")
	}
	return
}

func writeJSONContentType(w http.ResponseWriter) {
	w.Header().Set("content-type", "application/json")
}

func writeJSONResponse(w http.ResponseWriter, v interface{}) {
	writeJSONContentType(w)
	json.NewEncoder(w).Encode(v)
}
