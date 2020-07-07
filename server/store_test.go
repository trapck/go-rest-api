package server

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsertArticle(t *testing.T) {
	db := initDB(t)
	defer closeDB(t, db)

	inputArticle := SingleArticleHTTPWrap{Article{Title: "test insert article"}}
	outputArticle, err := db.CreateArticle(inputArticle)
	failOnNotEqual(t, err, nil, fmt.Sprintf("article must be created without error, instead got : %s", err))
	failOnEqual(t, "", outputArticle.Slug, "created article must have slug, but got empty string") //TODO: change slug to id
	foundNewArticle, err := db.GetArticle(outputArticle.Slug)
	failOnNotEqual(t, err, nil, fmt.Sprintf("expected to get just created article by slug value %q but got error. %q", outputArticle.Slug, err))
	assert.Equal(t, inputArticle.Title, foundNewArticle.Title, "new article should be found in db with the same title")
	//TODO: wrap into transaction to have clear db after test
}

func TestSelectArticle(t *testing.T) {
	db := initDB(t)
	defer closeDB(t, db)

	fakeSlug := "1 2 3 4 5"
	a, err := db.GetArticle(fakeSlug)
	failOnEqual(t, err, nil, fmt.Sprintf("expected to get error but found article %#v", a))
	// success test cases are covered in insert test
}

func initDB(t *testing.T) *DBBlogStore {
	t.Helper()
	db := DBBlogStore{}
	err := db.Init()
	if err != nil {
		assert.FailNow(t, "db connection was not established. ", err)
	}
	return &db
}

func closeDB(t *testing.T, db *DBBlogStore) {
	t.Helper()
	err := db.Close()
	if err != nil {
		assert.FailNow(t, "db connection was not closed. ", err)
	}
}

func failOnEqual(t *testing.T, v1 interface{}, v2 interface{}, msg string) {
	t.Helper()
	if v1 == v2 {
		assert.FailNow(t, msg)
	}
}

func failOnNotEqual(t *testing.T, v1 interface{}, v2 interface{}, msg string) {
	t.Helper()
	if v1 != v2 {
		assert.FailNow(t, msg)
	}
}
