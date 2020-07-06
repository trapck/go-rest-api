package server

import (
	"encoding/json"
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

// NewBlogServer initializes new instance of the blog server
func NewBlogServer(s BlogStore) *BlogServer {
	server := BlogServer{Store: s}
	router := http.NewServeMux()
	router.HandleFunc("/api/articles/", server.serveGetArticle)
	router.HandleFunc("/api/articles", server.serveCreateArticle)
	server.Handler = router
	return &server
}

func (s *BlogServer) serveGetArticle(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/api/articles/")
	article, err := s.Store.GetArticle(slug)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	} else {
		writeSuccessJSONResponse(&w, SingleArticleHTTPWrap{article})
	}
}

func (s *BlogServer) serveCreateArticle(w http.ResponseWriter, r *http.Request) {
	var reqData SingleArticleHTTPWrap
	json.NewDecoder(r.Body).Decode(&reqData)
	createdArticle, _ := s.Store.CreateArticle(reqData)
	writeSuccessJSONResponse(&w, createdArticle)
}

func writeSuccessJSONResponse(w *http.ResponseWriter, v interface{}) {
	(*w).Header().Set("content-type", "application/json")
	json.NewEncoder((*w)).Encode(v)
}
