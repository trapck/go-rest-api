package server

//Article is model of the blog article
type Article struct {
	Slug  string `db:"slug"`
	Title string `db:"title"`
}

//SingleArticleHTTPWrap is http request/response model for single article
type SingleArticleHTTPWrap struct {
	Article
}
