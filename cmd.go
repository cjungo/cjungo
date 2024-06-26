package cjungo

import (
	"reflect"

	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

// TODO 设计太差
type Command interface {
	Init(container DiContainer) error
	Exec(container DiContainer) error
}

type CommandRunner struct {
	container DiContainer
	commands  []Command
}

func NewCommand(commands ...Command) (*CommandRunner, error) {
	container := &DiSimpleContainer{
		Container: dig.New(),
	}
	// 日志
	if err := container.Provide(NewLogger); err != nil {
		return nil, err
	}

	return &CommandRunner{
		container: container,
		commands:  commands,
	}, nil
}

func (runner *CommandRunner) AddCommand(command Command) {
	runner.commands = append(runner.commands, command)
}

func getTypeName(v any) string {
	if t := reflect.TypeOf(v); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}

func (runner *CommandRunner) Run() error {
	return runner.container.Invoke(func(logger *zerolog.Logger) error {
		logger.Info().Str("action", "开始").Msg("[CMD]")
		// 初始化
		for _, command := range runner.commands {
			name := getTypeName(command)
			logger.Info().
				Str("action", "初始化").
				Str("name", name).
				Msg("[CMD]")
			if err := command.Init(runner.container); err != nil {
				return err
			}
		}

		// 调用
		for _, command := range runner.commands {
			name := getTypeName(command)
			logger.Info().
				Str("action", "执行").
				Str("name", name).
				Msg("[CMD]")
			if err := command.Exec(runner.container); err != nil {
				return err
			}
		}
		return nil
	})
}
