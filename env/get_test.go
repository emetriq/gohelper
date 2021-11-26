package env

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStrEnv(t *testing.T) {
	result := GetStrEnv("TEST_STR_ENV", "default")
	assert.NotNil(t, result)
	assert.Equal(t, result, "default")
	os.Setenv("TEST_STR_ENV", "mytest")
	result = GetStrEnv("TEST_STR_ENV", "default")
	assert.Equal(t, result, "mytest")
}

func TestGetIntEnv(t *testing.T) {
	result := GetIntEnv("TEST_STR_ENV", -2)
	assert.NotNil(t, result)
	assert.Equal(t, result, -2)
	os.Setenv("TEST_STR_ENV", "qdw")
	result = GetIntEnv("TEST_STR_ENV", -2)
	assert.Equal(t, result, -2)
	os.Setenv("TEST_STR_ENV", "1222")
	result = GetIntEnv("TEST_STR_ENV", -2)
	assert.Equal(t, result, 1222)
}
