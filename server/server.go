package server

import (
	"fmt"
	"net/http"
	"strings"
)

//BlogServer handles blog requests
func BlogServer(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/api/articles/")

	if slug == "some-art" {
		fmt.Fprint(w, "some art")
	}

	if slug == "some-other-art" {
		fmt.Fprint(w, "some other art")
	}

}
