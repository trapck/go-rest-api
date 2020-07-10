package server

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // db driver
)

// DBBlogStore is implementation of blog store via Postgres
type DBBlogStore struct {
	db *sqlx.DB
}

// Init initializes connetion
func (s *DBBlogStore) Init() (err error) {
	s.db, err = sqlx.Connect("postgres", "user=postgres password=postgres dbname=postgres sslmode=disable")
	return
}

// Close closes connetion
func (s *DBBlogStore) Close() (err error) {
	var isConnected bool
	if isConnected, err = s.ensureConnection(); isConnected {
		err = s.db.Close()
	}
	return
}

// GetArticle selects article from db by slug search value
func (s *DBBlogStore) GetArticle(slug string) (article Article, e error) {
	var a Article
	e = s.db.Get(&a, "SELECT * FROM article WHERE slug=$1", slug)
	return a, e
}

// CreateArticle creates article in db
func (s *DBBlogStore) CreateArticle(a SingleArticleHTTPWrap) (article Article, e error) {
	if isConnected, e := s.ensureConnection(); !isConnected {
		return article, e
	}
	a.Article.Slug = CreateSlug(a.Article.Title)
	_, err := s.db.NamedExec("INSERT INTO article (slug, title) VALUES (:slug, :title)", a.Article)
	return a.Article, err
}

// GetUser returns user from db
func (s *DBBlogStore) GetUser(username string) (RequestUserData, error) {
	var u RequestUserData
	e := s.db.Get(&u, "SELECT * FROM usr WHERE login=$1", username)
	return u, e
}

// Registration creates user in db
func (s *DBBlogStore) Registration(user RequestUserData) (RequestUserData, error) {
	if isConnected, e := s.ensureConnection(); !isConnected {
		return RequestUserData{}, e
	}
	_, err := s.db.NamedExec(`INSERT INTO usr (login, password, email, image, bio)
								VALUES (:login, :password, :email, :image, :bio)`, user)
	return user, err
}

func (s *DBBlogStore) ensureConnection() (isConnected bool, e error) {
	isConnected = s.db != nil
	if !isConnected {
		e = fmt.Errorf("db connection is not initialized")
	}
	return
}
