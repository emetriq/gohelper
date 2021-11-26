package collections

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapKeysStrStr(t *testing.T) {
	myMap := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	assert.ElementsMatch(t, []string{"key1", "key2"}, GetMapKeysStrStr(myMap))
}

func TestMapKeysStrInt(t *testing.T) {
	myMap := map[string]int{
		"key1": 1,
		"key2": 2,
	}
	assert.ElementsMatch(t, []string{"key1", "key2"}, GetMapKeysStrInt(myMap))
}

func TestMapKeysIntInt(t *testing.T) {
	myMap := map[int]int{
		1:  2,
		44: 3,
	}
	assert.ElementsMatch(t, []int{1, 44}, GetMapKeysIntInt(myMap))
}

func TestMapKeysIntStr(t *testing.T) {
	myMap := map[int]string{
		1:  "2",
		44: "3",
	}
	assert.ElementsMatch(t, []int{1, 44}, GetMapKeysIntStr(myMap))
}
