package server

// UnprocessableEntityResponse represents the response body for 422 responses
type UnprocessableEntityResponse struct {
	Errors UnprocessableEntityError
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
