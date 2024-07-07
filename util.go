package cjungo

import (
	"fmt"
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

func LimitStr(v string, n int) string {
	if len(v) > n {
		return fmt.Sprintf("%s...", v[:n-3])
	}
	return v
}
