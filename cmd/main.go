package main

import (
	"log"
	"net/http"

	"github.com/trapck/go-rest-api/server"
)

// InMemoryBlogStore stores blogs in memory
type InMemoryBlogStore struct{}

// GetArticle returns article from in memory blog store
func (s *InMemoryBlogStore) GetArticle(search string) string {
	return "article from fake store"
}

func main() {
	s := &server.BlogServer{Store: &InMemoryBlogStore{}}
	if err := http.ListenAndServe(":3000", s); err != nil {
		log.Fatalf("could not listen on port 3000 %v", err)
	}
}
