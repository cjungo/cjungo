package cjungo

import (
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func LoadEnv() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	envPath := filepath.Join(pwd, ".env")

	if IsFileExist(envPath) {
		if err := godotenv.Load(envPath); err != nil {
			return err
		}
	}

	return nil
}

func GetEnvInt[T int | uint16](name string, onResult func(T)) error {
	text := os.Getenv(name)
	if len(text) > 0 {
		v, err := strconv.Atoi(text)
		if err != nil {
			return err
		}
		r := T(v)
		onResult(r)
	}
	return nil
}

func GetEnvDuration(name string, onResult func(time.Duration)) error {
	text := os.Getenv(name)
	if len(text) > 0 {
		v, err := time.ParseDuration(text)
		if err != nil {
			return err
		}
		onResult(v)
	}
	return nil
}

func GetEnvBool(name string, onResult func(bool)) error {
	text := os.Getenv(name)
	if len(text) > 0 {
		v, err := strconv.ParseBool(text)
		if err != nil {
			return err
		}
		onResult(v)
	}
	return nil
}
