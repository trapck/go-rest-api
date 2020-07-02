package server

import (
	"fmt"
	"net/http"
	"strings"
)

// BlogStore stores blog data
type BlogStore interface {
	GetArticle(search string) string
}

// BlogServer handles bolg api requests
type BlogServer struct {
	Store BlogStore
}

func (s *BlogServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/api/articles/")
	fmt.Fprint(w, s.Store.GetArticle(slug))
}
