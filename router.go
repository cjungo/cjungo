package cjungo

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func NewRouter(logger *zerolog.Logger) *echo.Echo {
	router := echo.New()

	// 使用自定义上下文
	router.Use(ResetContext)

	// 异常处理句柄
	router.HTTPErrorHandler = func(err error, c echo.Context) {
		ctx := c.(HttpContext)
		code := http.StatusInternalServerError
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
		}
		tip := fmt.Sprintf("%d: %v", code, err)
		logger.Info().Str("error", tip).Msg("error:")
		ctx.RespBad(tip)
	}
	return router
}