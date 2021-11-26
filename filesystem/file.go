package filesystem

import (
	"errors"
	"os"
)

//ExistsWithError checks if a file or directory exists and returns error
func ExistsWithError(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

//Exists checks if a file or directory exists
func Exists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}
