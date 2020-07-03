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
}

func (s *BlogServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/articles" && r.Method == http.MethodPost {
		var reqData SingleArticleHTTPWrap
		_ = json.NewDecoder(r.Body).Decode(&reqData)
		r.Body.Close()
		createdArticle, _ := s.Store.CreateArticle(reqData)
		json.NewEncoder(w).Encode(SingleArticleHTTPWrap{createdArticle})
	} else {
		slug := strings.TrimPrefix(r.URL.Path, "/api/articles/")
		article, err := s.Store.GetArticle(slug)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
		} else {
			wrappedArticle := SingleArticleHTTPWrap{article}
			json.NewEncoder(w).Encode(wrappedArticle)
		}
	}
}
