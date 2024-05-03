package cjungo

import (
	"os"
)

func GetOrDefault[T any](v *T, d T) T {
	if v != nil {
		return *v
	}
	return d
}

func IsFileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
