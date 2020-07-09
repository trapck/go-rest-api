package server

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateToke(t *testing.T) {
	u1 := AuthData{"user1"}
	u2 := AuthData{"user2"}
	t.Run("must generate unique token per user", func(t *testing.T) {
		assert.NotEqual(t, CreateToken(u1), CreateToken(u2), fmt.Sprintf("have two equal tokens for %+v and %+v", u1, u2))
	})
}
