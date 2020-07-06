package main

import (
	"log"
	"net/http"

	"github.com/trapck/go-rest-api/server"
)

// InMemoryBlogStore stores blogs in memory
type InMemoryBlogStore struct{}

// GetArticle returns article from in memory blog store
func (s *InMemoryBlogStore) GetArticle(search string) (article server.Article, e error) {
	return server.Article{Slug: "found-article-from-fake-store", Title: "found article from fake store"}, nil
}

// CreateArticle returns article from in memory blog store
func (s *InMemoryBlogStore) CreateArticle(a server.SingleArticleHTTPWrap) (article server.Article, e error) {
	return server.Article{Slug: "created-fake-article", Title: "created fake article"}, nil
}

func main() {
	s := server.NewBlogServer(&InMemoryBlogStore{})
	if err := http.ListenAndServe(":3000", s); err != nil {
		log.Fatalf("could not listen on port 3000 %v", err)
	}
}
