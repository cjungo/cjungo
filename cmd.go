package cjungo

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

func RunCommand[T any](
	runner any,
	providers ...any,
) error {
	name := GetFuncName(runner)
	if strings.HasSuffix(name, ".main.func1") {
		if _, file, _, ok := runtime.Caller(1); ok {
			name = filepath.Base(filepath.Dir(file))
		} else {
			return fmt.Errorf("闭包任务获取目录名失败。")
		}
	}
	logPath := fmt.Sprintf("./log/%s.log", name)
	os.Setenv("CJUNGO_LOG_FILENAME", logPath)
	if err := LoadEnv(); err != nil {
		return err
	}

	container := &DiSimpleContainer{
		Container: dig.New(),
	}
	// 日志
	if err := container.Provides(NewLogger, LoadLoggerConfFromEnv); err != nil {
		return err
	}

	// 提供
	if err := container.Provides(providers...); err != nil {
		return err
	}

	// 参数
	if err := container.Provide(func() (*T, error) {
		var args T
		if _, err := flags.ParseArgs(&args, os.Args); err != nil {
			if fe, ok := err.(*flags.Error); ok && fe.Type == flags.ErrHelp {
				return &args, nil
			}
			return nil, err
		}
		return &args, nil
	}); err != nil {
		return err
	}

	return container.Invoke(func(logger *zerolog.Logger) error {
		logger.
			Info().
			Str("action", "开始").
			Str("name", name).
			Msg("[CMD]")
		if err := container.Invoke(runner); err != nil {
			return err
		}
		logger.
			Info().
			Str("action", "完成").
			Str("name", name).
			Msg("[CMD]")
		return nil
	})
}

func GetFuncName(v any) string {
	return runtime.FuncForPC(reflect.ValueOf(v).Pointer()).Name()
}

func GetTypeName(v any) string {
	if t := reflect.TypeOf(v); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}
