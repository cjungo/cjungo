package cjungo

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/dig"
)

type Application struct {
	container *dig.Container
	router    *echo.Echo
}

type ApplicationInitHandle func(container *dig.Container) error

func NewApplication(handle ApplicationInitHandle) (*Application, error) {
	container := dig.New()
	router := echo.New()
	router.HTTPErrorHandler = func(err error, ctx echo.Context) {
		code := http.StatusInternalServerError
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
		}
		ctx.JSON(
			http.StatusBadRequest,
			map[string]any{
				"code":    -1,
				"message": fmt.Sprintf("请求错误(%d)", code),
			},
		)
	}
	if err := handle(container); err != nil {
		return nil, err
	}
	if err := container.Provide(func() *echo.Echo { return router }); err != nil {
		return nil, err
	}
	if err := container.Provide(NewHttpServer); err != nil {
		return nil, err
	}
	return &Application{
		container: container,
		router:    router,
	}, nil
}

func (app *Application) Run() error {
	return app.container.Invoke(func(server *http.Server) error {
		if err := server.ListenAndServe(); err != nil {
			return err
		}
		return nil
	})
}
