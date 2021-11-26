package filesystem

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExistsWithError(t *testing.T) {
	ok, err := ExistsWithError("../README.md")
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = ExistsWithError("README_not_exists.md")
	assert.Nil(t, err)
	assert.False(t, ok)
}

func TestExists(t *testing.T) {
	ok := Exists("../README.md")
	assert.True(t, ok)

	ok = Exists("README_not_exists.md")
	assert.False(t, ok)
}
