// Package env provides utilities for loading and accessing environment variables.
package env

import (
	"os"
	"strconv"
)

// GetString returns the value of the environment variable identified by key,
// or fallback if the variable is not set.
func GetString(key, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	return val
}

// GetInt returns the value of the environment variable identified by key,
// or fallback if the variable is not set or if it's not an int type.
func GetInt(key string, fallback int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	valAsInt, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}

	return valAsInt
}
