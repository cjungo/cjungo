package cjungo

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
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
	if err := container.Provide(func(logger *zerolog.Logger) *echo.Echo {
		router := echo.New()
		router.HTTPErrorHandler = func(err error, ctx echo.Context) {
			code := http.StatusInternalServerError
			if he, ok := err.(*echo.HTTPError); ok {
				code = he.Code
			}
			logger.Info().Str("error", fmt.Sprintf("%v", err)).Msg("error:")
			ctx.JSON(
				http.StatusBadRequest,
				map[string]any{
					"code":    -1,
					"message": fmt.Sprintf("请求错误(%d)", code),
				},
			)
		}
		return router
	}); err != nil {
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
