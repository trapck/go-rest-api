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
func (s *DBBlogStore) GetArticle(slug string) (Article, error) {
	var a Article
	err := s.db.Get(&a, "SELECT * FROM article WHERE slug=$1", slug)
	if a.AuthorID.Valid && err == nil {
		var u RequestUserData
		if u, err = s.getUserByID(int(a.AuthorID.Int32)); err == nil {
			a.Author = u.ToProfile()
		}
	}
	return a, err
}

// CreateArticle creates article in db
func (s *DBBlogStore) CreateArticle(a SingleArticleHTTPWrap) (article Article, e error) {
	if isConnected, e := s.ensureConnection(); !isConnected {
		return article, e
	}
	a.Slug = CreateSlug(a.Title)
	_, err := s.db.Exec("INSERT INTO article (slug, title, author_id) VALUES ($1, $2, $3)", a.Slug, a.Title, a.AuthorID)
	if a.AuthorID.Valid && err == nil {
		if u, e := s.getUserByID(int(a.AuthorID.Int32)); e == nil {
			a.Author = u.ToProfile()
		}
	}
	return a.Article, err
}

// GetUser returns user from db
func (s *DBBlogStore) GetUser(username string) (RequestUserData, error) {
	var u RequestUserData
	e := s.db.Get(&u, "SELECT * FROM usr WHERE login=$1", username)
	return u, e
}

func (s *DBBlogStore) getUserByID(id int) (RequestUserData, error) {
	var u RequestUserData
	e := s.db.Get(&u, "SELECT * FROM usr WHERE id=$1", id)
	return u, e
}

// UpdateUser updates user in db
func (s *DBBlogStore) UpdateUser(username string, data RequestUserData) (RequestUserData, error) {
	_, err := s.db.Exec("UPDATE usr SET login=$1, password=$2, email=$3, bio=$4, image=$5 WHERE login=$6",
		data.UserName, data.Password, data.Email, data.Bio, data.Image, username)
	return data, err
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
