package env

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetHostname(t *testing.T) {
	result := GetHostname()
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)
}
