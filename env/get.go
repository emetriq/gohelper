package env

import (
	"os"
	"strconv"
)

func GetIntEnv(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		dig, err := strconv.Atoi(value)
		if err != nil {
			return fallback
		}
		return dig
	}
	return fallback
}

func GetStrEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
