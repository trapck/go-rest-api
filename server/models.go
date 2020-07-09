package server

import (
	"encoding/json"
)

// UnprocessableEntityResponse represents the response body for 422 responses
type UnprocessableEntityResponse struct {
	Errors UnprocessableEntityError
}

func (e *UnprocessableEntityResponse) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}

// UnprocessableEntityError represents error in 422 response body
type UnprocessableEntityError struct {
	Body []string
}

//Article is model of the blog article
type Article struct {
	Slug  string `db:"slug"`
	Title string `db:"title"`
}

//SingleArticleHTTPWrap is http request/response model for single article
type SingleArticleHTTPWrap struct {
	Article
}

// CommonUserData represents user data that is common for user request and response
type CommonUserData struct {
	Email    string
	UserName string
	Bio      string
	Image    string
}

// RequestUser represents user request data
type RequestUser struct {
	CommonUserData
	Password string
}

// ResponseUser represents user response data
type ResponseUser struct {
	CommonUserData
	Token string
}
