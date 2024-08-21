package cjungo

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"golang.org/x/exp/constraints"
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

func IsDirExist(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return stat.IsDir()
}

func GlobDir(path string) ([]string, error) {
	if !filepath.IsAbs(path) {
		if p, err := filepath.Abs(path); err != nil {
			return nil, err
		} else {
			path = p
		}
	}
	entities, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	result := []string{}
	for _, entity := range entities {
		p := filepath.Join(path, entity.Name())
		if entity.IsDir() {
			ps, err := GlobDir(p)
			if err != nil {
				return nil, err
			}
			result = append(result, ps...)
		} else {
			result = append(result, p)
		}
	}
	return result, nil
}

func LimitStr(v string, n int) string {
	if len(v) > n {
		return fmt.Sprintf("%s...", v[:n-3])
	}
	return v
}

func Max[T constraints.Ordered](a T, b T) T {
	if a > b {
		return a
	} else {
		return b
	}
}

func MaxOf[T constraints.Ordered](items ...T) T {
	var result *T
	for _, item := range items {
		if result == nil {
			result = &item
		} else if *result < item {
			result = &item
		}
	}
	return *result
}

func Min[T constraints.Ordered](a T, b T) T {
	if a < b {
		return a
	} else {
		return b
	}
}

func MinOf[T constraints.Ordered](items ...T) T {
	var result *T
	for _, item := range items {
		if result == nil {
			result = &item
		} else if *result > item {
			result = &item
		}
	}
	return *result
}

func Limit[T constraints.Ordered](min T, max T, v T) T {
	if v > max {
		return max
	} else if v < min {
		return min
	} else {
		return v
	}
}

func LimitOf[T constraints.Ordered](min T, max T, items ...T) []T {
	result := []T{}
	for _, item := range items {
		if item <= max && item >= min {
			result = append(result, item)
		}
	}
	return result
}

func IsNilOrEmpty(v *string) bool {
	return v == nil || *v == ""
}

func IsNilOrSpace(v *string) bool {
	return v == nil || strings.TrimSpace(*v) == ""
}

func IsNil(val any) bool {
	if val == nil {
		return true
	}
	v := reflect.ValueOf(val)
	k := v.Kind()
	switch k {
	case reflect.Chan,
		reflect.Func,
		reflect.Map,
		reflect.Pointer,
		reflect.UnsafePointer,
		reflect.Interface,
		reflect.Slice:
		return v.IsNil()
	}

	return false
}
