package server

import (
	"encoding/json"

	"github.com/dgrijalva/jwt-go"
)

// AuthData is data model to use for auth token
type AuthData struct {
	Login string
}

// AuthClaims is auth custom claims type
type AuthClaims struct {
	User AuthData `json:"user"`
	jwt.StandardClaims
}

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

// UpdateUserData is struct for update user request. Uses pointers to indicate null or json absent fields
type UpdateUserData struct {
	Email    *string
	UserName *string
	Bio      *string
	Image    *string
	Password *string
}

// UpdateUserRequest is request model to update user
type UpdateUserRequest struct {
	User UpdateUserData
}
