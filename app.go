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

	if err := container.Provides(
		NewLogger,     // 日志
		NewRouter,     // 路由
		NewHttpServer, // 服务器
	); err != nil {
		return nil, err
	}

	// 自定义
	if err := handle(container); err != nil {
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
			di.Logger.Info().Str("action", "启动队列...").Msg("[TASK]")
			err := di.Queue.Run()
			if err != nil {
				return err
			}
		} else {
			di.Logger.Info().Str("action", "没有启动队列").Msg("[TASK]")
		}
		di.Logger.Info().Str("action", "启动服务器").Msg("[HTTP]")
		return di.Server.ListenAndServe()
	})
}
