package server

import "strings"

// CreateSlug creates slug from title
func CreateSlug(title string) string {
	return strings.ToLower(strings.Join(strings.Fields(title), "-"))
}
