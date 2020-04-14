package utils

import (
	"errors"
	"os"
)

var ErrNoSuchKey = errors.New("no such env var")

func GetEnv(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return "", ErrNoSuchKey
	}
	return value, nil
}

func GetEnvDef(key, def string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	return value
}
