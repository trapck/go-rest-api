package server

// 422 error descriptions
const (
	MsgInvalidBody = "invalid json body"
)

// Auth depended constants
const (
	AuthHeader0Part = "Token"
	authSecretKey   = "qweasdzxc" // move to .env
)

// Constants for http header keys
const (
	HeaderKeyContentType   = "Content-Type"
	HeaderKeyAuthorization = "Authorization"
)

// Constants for http header values
const (
	HeaderValueJSONContactType = "application/json; charset=utf-8"
)
