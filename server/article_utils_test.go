package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateSlug(t *testing.T) {
	testCases := map[string]string{
		"test":                 "test",
		"test article":         "test-article",
		"   test":              "test",
		"test    ":             "test",
		"   test   article   ": "test-article",
		"TEST ARTICLE":         "test-article",
	}
	for title, slug := range testCases {
		assert.Equal(t, slug, CreateSlug(title))
	}
}
