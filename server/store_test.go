package server

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testUser = RequestUserData{CommonUserData: CommonUserData{ID: 47, UserName: "unit_test"}}

func TestInsertArticle(t *testing.T) {
	db := initDB(t)
	sessionID := createSessionID()
	defer closeDB(t, db)
	defer clearTestData(db, "article", fmt.Sprintf("title LIKE '%s'", "%"+sessionID+"%"))

	inputArticle := SingleArticleHTTPWrap{Article{Title: fmt.Sprintf("test%s insert article", sessionID), AuthorID: sql.NullInt32{Int32: int32(testUser.ID), Valid: true}}}
	outputArticle, err := db.CreateArticle(inputArticle)
	failOnNotEqual(t, err, nil, fmt.Sprintf("article must be created without error, instead got : %s", err))
	failOnEqual(t, "", outputArticle.Slug, "created article must have slug, but got empty string") //TODO: change slug to id
	assert.Equal(t, testUser.UserName, outputArticle.Author.UserName, "created article must have expected author")
	foundNewArticle, err := db.GetArticle(outputArticle.Slug)
	failOnNotEqual(t, err, nil, fmt.Sprintf("expected to get just created article by slug value %q but got error. %q", outputArticle.Slug, err))
	assert.Equal(t, inputArticle.Title, foundNewArticle.Title, "found in db new article should be found in db with the same title")
	assert.Equal(t, testUser.UserName, foundNewArticle.Author.UserName, "found in db new article must have expected author")
	// TODO: test duplicate rows
}

func TestSelectArticle(t *testing.T) {
	db := initDB(t)
	defer closeDB(t, db)

	fakeSlug := "1 2 3 4 5"
	a, err := db.GetArticle(fakeSlug)
	failOnEqual(t, err, nil, fmt.Sprintf("expected to get an error for search by fake slug %q but found article %#v", fakeSlug, a))
	// success test cases are covered in insert test
}

func TestInsertUser(t *testing.T) {
	db := initDB(t)
	sessionID := createSessionID()
	defer closeDB(t, db)
	defer clearTestData(db, "usr", fmt.Sprintf("login LIKE '%s'", "%"+sessionID+"%"))

	inputUser := RequestUserData{
		CommonUserData: CommonUserData{
			UserName: fmt.Sprintf("test%s registration", sessionID),
			Email:    "registration@gmail.com",
		},
		Password: "123",
	}
	outputUser, err := db.Registration(inputUser)
	failOnNotEqual(t, err, nil, fmt.Sprintf("user must be created without error, instead got : %s", err))
	failOnNotEqual(
		t,
		inputUser.UserName,
		outputUser.UserName,
		fmt.Sprintf("created user must have username %q, but got %q", inputUser.UserName, outputUser.UserName),
	) //TODO: change UserName to id
	foundNewUser, err := db.GetUser(outputUser.UserName)
	failOnNotEqual(t, err, nil, fmt.Sprintf("expected to get just created user by username value %q but got error. %q", outputUser.UserName, err))
	assert.Equal(t, inputUser.Email, foundNewUser.Email, "new user should be found in db with the same email")

	// TODO: test duplicate users
}

func TestUpdateUser(t *testing.T) {
	db := initDB(t)
	sessionID := createSessionID()
	currentUserName := "test_update_" + sessionID
	defer closeDB(t, db)
	defer clearTestData(db, "usr", fmt.Sprintf("login LIKE '%s'", "%"+sessionID+"%"))
	_, e := db.db.Exec("INSERT INTO usr (login) VALUES ($1)", currentUserName)
	failOnNotEqual(t, e, nil, fmt.Sprintf("could not insert test data into db. %q", e))

	updateData := RequestUserData{CommonUserData: CommonUserData{UserName: currentUserName + "_updated", Email: "e", Bio: "b", Image: "i"}, Password: "p"}
	updatedUser, e := db.UpdateUser(currentUserName, updateData)

	failOnNotEqual(t, e, nil, fmt.Sprintf("expected to update user without errors but got %q", e))
	assert.Equal(t, updateData, updatedUser, "expected returned user to be equal to input data")

	e = db.db.Get(&updatedUser, "SELECT * FROM usr WHERE login=$1", updateData.UserName)
	failOnNotEqual(t, e, nil, fmt.Sprintf("expected to select updated user from db without errors but got %q", e))
	updateData.ID = updatedUser.ID
	assert.Equal(t, updateData, updatedUser, "expected to find user with updated data in db")

	// TODO: test duplicate users
}

func TestSelectUser(t *testing.T) {
	db := initDB(t)
	defer closeDB(t, db)

	fakeUserName := "user1 user2 user3 user4 user5"
	a, err := db.GetUser(fakeUserName)
	failOnEqual(t, err, nil, fmt.Sprintf("expected to get an error for search by fake username %q but found users %#v", fakeUserName, a))
	// success test cases are covered in insert user test
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

func clearTestData(db *DBBlogStore, table, filter string) {
	db.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE %s", table, filter))
}

func createSessionID() string {
	return strconv.Itoa(int(rand.Uint32()))
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
