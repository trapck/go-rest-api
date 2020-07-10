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

// Article is model of the blog article
type Article struct {
	Slug  string `db:"slug"`
	Title string `db:"title"`
}

// SingleArticleHTTPWrap is http request/response model for single article
type SingleArticleHTTPWrap struct {
	Article
}

// CommonUserData represents user data that is common for user request and response
type CommonUserData struct {
	Email    string `db:"email"`
	UserName string `db:"login"`
	Bio      string `db:"bio"`
	Image    string `db:"image"`
}

// RequestUserData represents user request data
type RequestUserData struct {
	CommonUserData
	Password string `db:"password"`
}

// ToCommonUserData converts current type to CommonUserData
func (u *RequestUserData) ToCommonUserData() CommonUserData {
	return CommonUserData{
		Email:    u.Email,
		UserName: u.UserName,
		Bio:      u.Bio,
		Image:    u.Image,
	}
}

// ResponseUserData represents user response data
type ResponseUserData struct {
	CommonUserData
	Token string
}

// RequestUser is user http request model
type RequestUser struct {
	User RequestUserData
}

// ResponseUser is user http response model
type ResponseUser struct {
	User ResponseUserData
}
