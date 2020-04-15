package utils

import (
	"errors"
	"os"
)

// ErrNoSuchKey no such env var
var ErrNoSuchKey = errors.New("no such env var")

// GetEnv get env var or error if not found
func GetEnv(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return "", ErrNoSuchKey
	}
	return value, nil
}

// GetEnvDef get env var or default value if not found
func GetEnvDef(key, def string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	return value
}
