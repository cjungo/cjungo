package cjungo

import (
	"net/http"

	"go.uber.org/dig"
)

type Application struct {
	container DiContainer
}

type ApplicationInitHandle func(container DiContainer) error

func NewApplication(handle ApplicationInitHandle) (*Application, error) {
	container := &DiSimpleContainer{
		Container: *dig.New(),
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
	}, nil
}

func (app *Application) Run() error {
	if err := app.container.Invoke(func(queue *TaskQueue) error {
		return queue.Run()
	}); err != nil {
		return err
	}

	return app.container.Invoke(func(server *http.Server) error {
		if err := server.ListenAndServe(); err != nil {
			return err
		}
		return nil
	})
}
