package collections

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapKeysStr(t *testing.T) {
	myMap := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	assert.Equal(t, []string{"key1", "key2"}, GetMapKeysStr(myMap))
}

func TestMapKeysInt(t *testing.T) {
	myMap := map[int]int{
		1:  2,
		44: 3,
	}
	assert.Equal(t, []int{1, 44}, GetMapKeysInt(myMap))
}
