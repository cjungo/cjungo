package cjungo

import (
	"net/http"

	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

type Application struct {
	container DiContainer
	BeforeRun func(DiContainer) error
}

type ApplicationInitHandle func(container DiContainer) error

func NewApplication(handle ApplicationInitHandle) (*Application, error) {
	container := &DiSimpleContainer{
		Container: dig.New(),
	}
	// 日志
	if err := container.Provide(NewLogger); err != nil {
		return nil, err
	}

	// 自定义
	if err := handle(container); err != nil {
		return nil, err
	}

	// 路由
	if err := container.Provide(NewRouter); err != nil {
		return nil, err
	}

	// 服务器
	if err := container.Provide(NewHttpServer); err != nil {
		return nil, err
	}

	return &Application{
		container: container,
		BeforeRun: func(_ DiContainer) error { return nil },
	}, nil
}

type ApplicationRunDi struct {
	dig.In
	Logger *zerolog.Logger
	Server *http.Server
	Queue  *TaskQueue `optional:"true"`
}

func (app *Application) Run() error {
	return app.container.Invoke(func(di ApplicationRunDi) error {
		// 前切入点
		if err := app.BeforeRun(app.container); err != nil {
			return err
		}

		// 队列服务
		if di.Queue != nil {
			di.Logger.Info().Msg("启动队列...")
			err := di.Queue.Run()
			if err != nil {
				return err
			}
		} else {
			di.Logger.Info().Msg("没有启动队列")
		}
		di.Logger.Info().Msg("启动服务器")
		return di.Server.ListenAndServe()
	})
}
